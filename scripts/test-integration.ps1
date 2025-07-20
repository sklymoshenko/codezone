#!/usr/bin/env pwsh

# PostgreSQL Integration Test Runner for Windows

param(
    [string]$Command = "all",
    [string]$Version = "15",
    [string]$Port = "5433",
    [switch]$Keep,
    [switch]$Recreate,
    [switch]$Help
)

# Colors for output
$Red = "`e[31m"
$Green = "`e[32m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Reset = "`e[0m"

function Write-ColorOutput {
    param(
        [string]$Color,
        [string]$Message
    )
    Write-Host "$Color$Message$Reset"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput $Blue "[INFO] $Message"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput $Green "[SUCCESS] $Message"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput $Yellow "[WARNING] $Message"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput $Red "[ERROR] $Message"
}

# Configuration
$ContainerName = "codezone-postgres-test"
$PostgresVersion = $Version
$PostgresPassword = "testpassword"
$PostgresDB = "testdb"
$PostgresUser = "testuser"
$PostgresPort = $Port
$MaxWaitTime = 60

function Test-Docker {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error "Docker is not installed or not in PATH"
        Write-Error "Please install Docker to run integration tests"
        exit 1
    }
    
    try {
        docker info | Out-Null
    }
    catch {
        Write-Error "Docker daemon is not running"
        Write-Error "Please start Docker and try again"
        exit 1
    }
    
    Write-Success "Docker is available"
}

function Test-ContainerExists {
    $containers = docker ps -a --format '{{.Names}}' 2>$null
    return $containers -contains $ContainerName
}

function Test-ContainerRunning {
    $containers = docker ps --format '{{.Names}}' 2>$null
    return $containers -contains $ContainerName
}

function Test-ContainerHealthy {
    if (Test-ContainerRunning) {
        docker exec $ContainerName pg_isready -U $PostgresUser -d $PostgresDB 2>$null
        return $LASTEXITCODE -eq 0
    }
    return $false
}

function Start-Postgres {
    Write-Info "Starting PostgreSQL $PostgresVersion container..."
    
    # Remove existing container if it exists
    if (Test-ContainerExists) {
        Write-Info "Removing existing container: $ContainerName"
        docker rm -f $ContainerName 2>$null
    }
    
    # Create a named volume for this test session
    $TestVolumeName = "$ContainerName-data"
    Write-Info "Creating named volume: $TestVolumeName"
    docker volume create $TestVolumeName | Out-Null
    
    $containerId = docker run --name $ContainerName `
        -e POSTGRES_PASSWORD=$PostgresPassword `
        -e POSTGRES_DB=$PostgresDB `
        -e POSTGRES_USER=$PostgresUser `
        -p "127.0.0.1:$PostgresPort`:5432" `
        -v "$TestVolumeName`:/var/lib/postgresql/data" `
        -d "postgres:$PostgresVersion"
    
    Write-Success "PostgreSQL container started with named volume: $TestVolumeName"
}

function Wait-ForPostgres {
    Write-Info "Waiting for PostgreSQL to be ready..."
    
    for ($i = 1; $i -le $MaxWaitTime; $i++) {
        if (docker exec $ContainerName pg_isready -U $PostgresUser -d $PostgresDB 2>$null) {
            Write-Success "PostgreSQL is ready!"
            
            # Additional wait to ensure PostgreSQL is fully ready for connections
            Write-Info "Waiting additional 3 seconds for PostgreSQL to be fully ready..."
            Start-Sleep 3
            
            return $true
        }
        
        if ($i -eq $MaxWaitTime) {
            Write-Error "PostgreSQL failed to start within $MaxWaitTime seconds"
            return $false
        }
        
        Write-Host "." -NoNewline
        Start-Sleep 1
    }
}

function Ensure-PostgresRunning {
    if (Test-ContainerExists) {
        if (Test-ContainerRunning) {
            if (Test-ContainerHealthy) {
                Write-Success "PostgreSQL container is already running and healthy"
                return $true
            }
            else {
                Write-Warning "Container exists but is not healthy, restarting..."
                docker stop $ContainerName 2>$null
                docker start $ContainerName 2>$null
            }
        }
        else {
            Write-Info "Container exists but is stopped, starting..."
            docker start $ContainerName 2>$null
        }
        
        if (Wait-ForPostgres) {
            return $true
        }
        else {
            Write-Warning "Failed to start existing container, recreating..."
            docker rm $ContainerName 2>$null
        }
    }
    
    Start-Postgres
    Wait-ForPostgres
}

function Remove-Container {
    if (Test-ContainerExists) {
        # Get volumes associated with this specific container before stopping it
        Write-Info "Finding volumes associated with container: $ContainerName"
        $containerVolumes = docker inspect $ContainerName --format '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' 2>$null
        
        Write-Info "Stopping and removing container: $ContainerName"
        docker stop $ContainerName 2>$null
        docker rm $ContainerName 2>$null
        
        # Remove our specific test volume
        $TestVolumeName = "$ContainerName-data"
        if (docker volume ls -q | Select-String -Pattern "^$TestVolumeName$") {
            Write-Info "Removing test volume: $TestVolumeName"
            docker volume rm $TestVolumeName 2>$null
        }
        
        # Remove any other volumes that were associated with our test container
        if ($containerVolumes) {
            Write-Info "Removing other volumes associated with test container: $containerVolumes"
            $volumes = $containerVolumes -split ' '
            foreach ($volume in $volumes) {
                # Skip our named volume as we already handled it above
                if ($volume -ne $TestVolumeName) {
                    Write-Info "Removing volume: $volume"
                    docker volume rm $volume 2>$null
                }
            }
        }
        else {
            Write-Info "No additional volumes found associated with test container"
        }
    }
}

function Remove-Complete {
    Remove-Container
    
    # Remove the PostgreSQL image to free up space
    Write-Info "Removing PostgreSQL $PostgresVersion image..."
    docker rmi "postgres:$PostgresVersion" 2>$null
    
    # Clean up any dangling images
    Write-Info "Cleaning up dangling Docker images..."
    docker image prune -f 2>$null
    
    # Only remove truly unused volumes (not all postgres volumes)
    Write-Info "Removing truly unused Docker volumes..."
    $unusedVolumeCount = (docker volume ls -q --filter dangling=true | Measure-Object -Line).Lines
    if ($unusedVolumeCount -gt 0) {
        Write-Info "Found $unusedVolumeCount unused (dangling) volumes to clean up"
        docker volume prune -f 2>$null
    }
    else {
        Write-Info "No unused volumes found"
    }
    
    Write-Success "Complete cleanup finished - only test-related resources removed"
}



function Run-IntegrationTests {
    Write-Info "Running integration tests..."
    
    $env:POSTGRES_HOST = "localhost"
    $env:POSTGRES_PORT = $PostgresPort
    $env:POSTGRES_DB = $PostgresDB
    $env:POSTGRES_USER = $PostgresUser
    $env:POSTGRES_PASSWORD = $PostgresPassword
    $env:CODEZONE_TEST_MODE = "true"
    
    $testOutput = go test ./executor -v -run "TestPostgreSQLExecutor_(Integration|ConnectionTesting)" -timeout=60s -count=1 2>&1
    $testExitCode = $LASTEXITCODE
    
    # Display the test output
    Write-Host $testOutput
    
    if ($testExitCode -eq 0) {
        Write-Success "Integration tests passed!"
        return $true
    }
    else {
        Write-Error "Integration tests failed!"
        return $false
    }
}

function Show-ConnectionInfo {
    Write-Info "PostgreSQL connection details:"
    Write-Host "  Host: localhost"
    Write-Host "  Port: $PostgresPort"
    Write-Host "  Database: $PostgresDB"
    Write-Host "  User: $PostgresUser"
    Write-Host "  Password: $PostgresPassword"
    Write-Host ""
    Write-Host "Connect with: psql -h localhost -p $PostgresPort -U $PostgresUser -d $PostgresDB"
}

function Show-Usage {
    Write-Host "Usage: $($MyInvocation.MyCommand.Name) [OPTIONS] [COMMAND]"
    Write-Host ""
    Write-Host "Commands:"
    Write-Host "  integration   Run only integration tests (requires database)"
    Write-Host "  all           Run all integration tests (default)"
    Write-Host "  setup         Start PostgreSQL container only"
    Write-Host "  cleanup       Stop and remove PostgreSQL container"
    Write-Host "  cleanup-all   Stop container, remove container and image"
    Write-Host "  logs          Show PostgreSQL container logs"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Version VERSION    PostgreSQL version (default: 15)"
    Write-Host "  -Port PORT         PostgreSQL port (default: 5433)"
    Write-Host "  -Keep              Keep container running after tests"
    Write-Host "  -Recreate          Force recreate container even if healthy"
    Write-Host "  -Help              Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  $($MyInvocation.MyCommand.Name)                      # Run all integration tests (reuse existing container)"
    Write-Host "  $($MyInvocation.MyCommand.Name) integration          # Run only integration tests"
    Write-Host "  $($MyInvocation.MyCommand.Name) setup                # Start PostgreSQL for manual testing"
    Write-Host "  $($MyInvocation.MyCommand.Name) -Version 16 all      # Test against PostgreSQL 16"
    Write-Host "  $($MyInvocation.MyCommand.Name) -Keep integration    # Keep container after integration tests"
    Write-Host "  $($MyInvocation.MyCommand.Name) -Recreate all        # Force recreate container"
}

# Main execution
function Main {
    Write-Info "PostgreSQL Integration Test Runner"
    Write-Info "Command: $Command"
    Write-Host ""
    
    switch ($Command) {
        "cleanup" {
            Remove-Container
            Write-Success "Container cleanup completed"
        }
        
        "cleanup-all" {
            Remove-Complete
            Write-Success "Complete cleanup finished"
        }
        
        "logs" {
            if (Test-ContainerRunning) {
                docker logs $ContainerName
            }
            else {
                Write-Error "Container $ContainerName is not running"
                exit 1
            }
        }
        

        
        "setup" {
            Test-Docker
            if ($Recreate) {
                Remove-Container
                Start-Postgres
                Wait-ForPostgres
            }
            else {
                Ensure-PostgresRunning
            }
            Show-ConnectionInfo
            Write-Success "PostgreSQL is ready for testing!"
            Write-Info "Run '$($MyInvocation.MyCommand.Name) cleanup' to stop the container when done"
        }
        
        "integration" {
            Test-Docker
            if ($Recreate) {
                Remove-Container
                Start-Postgres
                Wait-ForPostgres
            }
            else {
                Ensure-PostgresRunning
            }
            
            Show-ConnectionInfo
            Write-Host ""
            
            if (Run-IntegrationTests) {
                Write-Success "All integration tests completed successfully!"
                $ExitCode = 0
            }
            else {
                Write-Error "Integration tests failed!"
                $ExitCode = 1
            }
            
            if (-not $Keep) {
                Write-Info "Stopping container (use -Keep to keep running)"
                docker stop $ContainerName 2>$null
            }
            else {
                Write-Info "Container kept running (use '$($MyInvocation.MyCommand.Name) cleanup' to remove)"
            }
            
            exit $ExitCode
        }
        
        "all" {
            Test-Docker
            
            if ($Recreate) {
                Remove-Container
                Start-Postgres
                Wait-ForPostgres
            }
            else {
                Ensure-PostgresRunning
            }
            
            Show-ConnectionInfo
            Write-Host ""
            
            if (Run-IntegrationTests) {
                Write-Success "All integration tests completed successfully!"
                $ExitCode = 0
            }
            else {
                Write-Error "Integration tests failed!"
                $ExitCode = 1
            }
            
            if (-not $Keep) {
                Write-Info "Stopping container (use -Keep to keep running)"
                docker stop $ContainerName 2>$null
            }
            else {
                Write-Info "Container kept running (use '$($MyInvocation.MyCommand.Name) cleanup' to remove)"
            }
            
            exit $ExitCode
        }
        
        default {
            Write-Error "Unknown command: $Command"
            Show-Usage
            exit 1
        }
    }
}

# Handle help parameter
if ($Help) {
    Show-Usage
    exit 0
}

# Only cleanup on unexpected exits, not normal completion
trap {
    if ($LASTEXITCODE -ne 0 -and -not $Keep -and $Command -ne "setup" -and $Command -ne "cleanup" -and $Command -ne "cleanup-all" -and $Command -ne "logs") {
        docker stop $ContainerName 2>$null
    }
}

# Run main function
Main 
#!/bin/bash

# PostgreSQL Integration Test Runner

set -e 

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

CONTAINER_NAME="codezone-postgres-test"
POSTGRES_VERSION="15"
POSTGRES_PASSWORD="testpassword"
POSTGRES_DB="testdb"
POSTGRES_USER="testuser"
POSTGRES_PORT="5433" # Changed from 5432 to 5433 to avoid conflicts with other services
MAX_WAIT_TIME=30

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        log_error "Please install Docker to run integration tests"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        log_error "Please start Docker and try again"
        exit 1
    fi
    
    log_success "Docker is available"
}

container_exists() {
    docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
}

container_running() {
    docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"
}

container_healthy() {
    if container_running; then
        docker exec "${CONTAINER_NAME}" pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" &> /dev/null
    else
        return 1
    fi
}

ensure_postgres_running() {
    if container_exists; then
        if container_running; then
            if container_healthy; then
                log_success "PostgreSQL container is already running and healthy"
                return 0
            else
                log_warning "Container exists but is not healthy, restarting..."
                docker stop "${CONTAINER_NAME}" &> /dev/null || true
                docker start "${CONTAINER_NAME}" &> /dev/null
            fi
        else
            log_info "Container exists but is stopped, starting..."
            docker start "${CONTAINER_NAME}" &> /dev/null
        fi
        
        if wait_for_postgres; then
            return 0
        else
            log_warning "Failed to start existing container, recreating..."
            docker rm "${CONTAINER_NAME}" &> /dev/null || true
        fi
    fi
    
    start_postgres
    wait_for_postgres
}

cleanup_container() {
    if container_exists; then
        # Get volumes associated with this specific container before stopping it
        log_info "Finding volumes associated with container: ${CONTAINER_NAME}"
        container_volumes=$(docker inspect "${CONTAINER_NAME}" --format '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}} {{end}}{{end}}' 2>/dev/null || echo "")
        
        log_info "Stopping and removing container: ${CONTAINER_NAME}"
        docker stop "${CONTAINER_NAME}" &> /dev/null || true
        docker rm "${CONTAINER_NAME}" &> /dev/null || true
        
        # Remove our specific test volume
        TEST_VOLUME_NAME="${CONTAINER_NAME}-data"
        if docker volume ls -q | grep -q "^${TEST_VOLUME_NAME}$"; then
            log_info "Removing test volume: ${TEST_VOLUME_NAME}"
            docker volume rm "${TEST_VOLUME_NAME}" &> /dev/null || true
        fi
        
        # Remove any other volumes that were associated with our test container
        if [ -n "$container_volumes" ]; then
            log_info "Removing other volumes associated with test container: $container_volumes"
            for volume in $container_volumes; do
                # Skip our named volume as we already handled it above
                if [ "$volume" != "${TEST_VOLUME_NAME}" ]; then
                    log_info "Removing volume: $volume"
                    docker volume rm "$volume" &> /dev/null || true
                fi
            done
        else
            log_info "No additional volumes found associated with test container"
        fi
    fi
}

cleanup_complete() {
    cleanup_container
    
    # Remove the PostgreSQL image to free up space
    log_info "Removing PostgreSQL ${POSTGRES_VERSION} image..."
    docker rmi "postgres:${POSTGRES_VERSION}" &> /dev/null || true
    
    # Clean up any dangling images
    log_info "Cleaning up dangling Docker images..."
    docker image prune -f &> /dev/null || true
    
    # Only remove truly unused volumes (not all postgres volumes)
    log_info "Removing truly unused Docker volumes..."
    unused_volume_count=$(docker volume ls -q --filter dangling=true | wc -l)
    if [ "$unused_volume_count" -gt 0 ]; then
        log_info "Found $unused_volume_count unused (dangling) volumes to clean up"
        docker volume prune -f &> /dev/null || true
    else
        log_info "No unused volumes found"
    fi
    
    log_success "Complete cleanup finished - only test-related resources removed"
}

start_postgres() {
    log_info "Starting PostgreSQL ${POSTGRES_VERSION} container..."
    
    # Create a named volume for this test session
    TEST_VOLUME_NAME="${CONTAINER_NAME}-data"
    log_info "Creating named volume: ${TEST_VOLUME_NAME}"
    docker volume create "${TEST_VOLUME_NAME}" > /dev/null
    
    docker run --name "${CONTAINER_NAME}" \
        -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
        -e POSTGRES_DB="${POSTGRES_DB}" \
        -e POSTGRES_USER="${POSTGRES_USER}" \
        -p "${POSTGRES_PORT}:5432" \
        -v "${TEST_VOLUME_NAME}:/var/lib/postgresql/data" \
        -d postgres:${POSTGRES_VERSION} \
        > /dev/null
    
    log_success "PostgreSQL container started with named volume: ${TEST_VOLUME_NAME}"
}

wait_for_postgres() {
    log_info "Waiting for PostgreSQL to be ready..."
    
    for i in $(seq 1 $MAX_WAIT_TIME); do
        if docker exec "${CONTAINER_NAME}" pg_isready -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" &> /dev/null; then
            log_success "PostgreSQL is ready!"
            return 0
        fi
        
        if [ $i -eq $MAX_WAIT_TIME ]; then
            log_error "PostgreSQL failed to start within ${MAX_WAIT_TIME} seconds"
            return 1
        fi
        
        echo -n "."
        sleep 1
    done
}

run_unit_tests() {
    log_info "Running unit tests..."
    
    if go test ./executor -v \
        -run "TestPostgreSQLExecutor_(Basic|Configuration|QueryType|IsSelect|PrepareSQLCode|ConvertValue|ExecuteWithoutConnection)" \
        -timeout=30s; then
        log_success "Unit tests passed!"
        return 0
    else
        log_error "Unit tests failed!"
        return 1
    fi
}

run_integration_tests() {
    log_info "Running integration tests..."
    
    export POSTGRES_HOST="localhost"
    export POSTGRES_PORT="${POSTGRES_PORT}"
    export POSTGRES_DB="${POSTGRES_DB}"
    export POSTGRES_USER="${POSTGRES_USER}"
    export POSTGRES_PASSWORD="${POSTGRES_PASSWORD}"
    
    if go test ./executor -v \
        -run "TestPostgreSQLExecutor_(Integration|ConnectionTesting)" \
        -timeout=60s; then
        log_success "Integration tests passed!"
        return 0
    else
        log_error "Integration tests failed!"
        return 1
    fi
}

show_connection_info() {
    log_info "PostgreSQL connection details:"
    echo "  Host: localhost"
    echo "  Port: ${POSTGRES_PORT}"
    echo "  Database: ${POSTGRES_DB}"
    echo "  User: ${POSTGRES_USER}"
    echo "  Password: ${POSTGRES_PASSWORD}"
    echo ""
    echo "Connect with: psql -h localhost -p ${POSTGRES_PORT} -U ${POSTGRES_USER} -d ${POSTGRES_DB}"
}

print_usage() {
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit          Run only unit tests (no database required)"
    echo "  integration   Run only integration tests (requires database)"
    echo "  all           Run all tests (default)"
    echo "  setup         Start PostgreSQL container only"
    echo "  cleanup       Stop and remove PostgreSQL container"
    echo "  cleanup-all   Stop container, remove container and image"
    echo "  logs          Show PostgreSQL container logs"
    echo ""
    echo "Options:"
    echo "  -v, --version VERSION    PostgreSQL version (default: 15)"
    echo "  -p, --port PORT         PostgreSQL port (default: 5432)"
    echo "  -k, --keep              Keep container running after tests"
    echo "  -r, --recreate          Force recreate container even if healthy"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run all tests (reuse existing container)"
    echo "  $0 unit                 # Run only unit tests"
    echo "  $0 integration          # Run only integration tests"
    echo "  $0 setup                # Start PostgreSQL for manual testing"
    echo "  $0 -v 16 all            # Test against PostgreSQL 16"
    echo "  $0 -k integration       # Keep container after integration tests"
    echo "  $0 -r all               # Force recreate container"
}

# Parse command line arguments
COMMAND="all"
KEEP_CONTAINER=false
FORCE_RECREATE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            POSTGRES_VERSION="$2"
            shift 2
            ;;
        -p|--port)
            POSTGRES_PORT="$2"
            shift 2
            ;;
        -k|--keep)
            KEEP_CONTAINER=true
            shift
            ;;
        -r|--recreate)
            FORCE_RECREATE=true
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        unit|integration|all|setup|cleanup|cleanup-all|logs)
            COMMAND="$1"
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    log_info "PostgreSQL Integration Test Runner"
    log_info "Command: ${COMMAND}"
    echo ""
    
    case "${COMMAND}" in
        cleanup)
            cleanup_container
            log_success "Container cleanup completed"
            ;;
            
        cleanup-all)
            cleanup_complete
            log_success "Complete cleanup finished"
            ;;
            
        logs)
            if container_running; then
                docker logs "${CONTAINER_NAME}"
            else
                log_error "Container ${CONTAINER_NAME} is not running"
                exit 1
            fi
            ;;
            
        unit)
            check_docker  # Still check docker for consistency
            run_unit_tests
            ;;
            
        setup)
            check_docker
            if [ "$FORCE_RECREATE" = true ]; then
                cleanup_container
                start_postgres
                wait_for_postgres
            else
                ensure_postgres_running
            fi
            show_connection_info
            log_success "PostgreSQL is ready for testing!"
            log_info "Run '$0 cleanup' to stop the container when done"
            ;;
            
        integration)
            check_docker
            if [ "$FORCE_RECREATE" = true ]; then
                cleanup_container
                start_postgres
                wait_for_postgres
            else
                ensure_postgres_running
            fi
            
            show_connection_info
            echo ""
            
            if run_integration_tests; then
                log_success "All integration tests completed successfully!"
                EXIT_CODE=0
            else
                log_error "Integration tests failed!"
                EXIT_CODE=1
            fi
            
            if [ "$KEEP_CONTAINER" = false ]; then
                log_info "Stopping container (use -k to keep running)"
                docker stop "${CONTAINER_NAME}" &> /dev/null || true
            else
                log_info "Container kept running (use '$0 cleanup' to remove)"
            fi
            
            exit $EXIT_CODE
            ;;
            
        all)
            check_docker
            
            # Run unit tests first (fast feedback)
            if ! run_unit_tests; then
                log_error "Unit tests failed, skipping integration tests"
                exit 1
            fi
            
            echo ""
            if [ "$FORCE_RECREATE" = true ]; then
                cleanup_container
                start_postgres
                wait_for_postgres
            else
                ensure_postgres_running
            fi
            
            show_connection_info
            echo ""
            
            if run_integration_tests; then
                log_success "All tests completed successfully!"
                EXIT_CODE=0
            else
                log_error "Integration tests failed!"
                EXIT_CODE=1
            fi
            
            if [ "$KEEP_CONTAINER" = false ]; then
                log_info "Stopping container (use -k to keep running)"
                docker stop "${CONTAINER_NAME}" &> /dev/null || true
            else
                log_info "Container kept running (use '$0 cleanup' to remove)"
            fi
            
            exit $EXIT_CODE
            ;;
            
        *)
            log_error "Unknown command: ${COMMAND}"
            print_usage
            exit 1
            ;;
    esac
}

# Only cleanup on unexpected exits, not normal completion
trap 'if [ $? -ne 0 ] && [ "$KEEP_CONTAINER" = false ] && [ "$COMMAND" != "setup" ] && [ "$COMMAND" != "cleanup" ] && [ "$COMMAND" != "cleanup-all" ] && [ "$COMMAND" != "logs" ]; then docker stop "${CONTAINER_NAME}" &> /dev/null || true; fi' EXIT

# Run main function
main 
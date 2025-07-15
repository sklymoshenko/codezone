import { Component, createSignal, Show } from 'solid-js'
import { SetPostgreSQLConfig, TestPostgreSQLConnection } from 'wailsjs/go/main/App'
import { executor } from 'wailsjs/go/models'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from './dialog'
import { showErrorToast } from './ErrorToast'
import { TextField, TextFieldInput, TextFieldLabel } from './text-field'

interface PostgresConnectionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConnect: (connected: boolean) => void
}

const PostgresConnectionDialog: Component<PostgresConnectionDialogProps> = props => {
  const [formData, setFormData] = createSignal({
    host: 'localhost',
    port: 5432,
    database: '',
    username: '',
    password: '',
    sslMode: 'disable'
  })

  const [isConnecting, setIsConnecting] = createSignal(false)

  const handleConnect = async () => {
    setIsConnecting(true)

    try {
      const config = new executor.PostgreSQLConfig(formData())

      // Test connection first
      const isValid = await TestPostgreSQLConnection(config)

      if (isValid) {
        // Set config in backend
        await SetPostgreSQLConfig(config)

        // Update connection status
        props.onConnect(true)
        props.onOpenChange(false)

        // Store config in localStorage for persistence (without password)
        const { password, ...configWithoutPassword } = formData()
        localStorage.setItem('postgres-config', JSON.stringify(configWithoutPassword))
      } else {
        showErrorToast({
          title: 'Connection Failed',
          description:
            'Failed to connect to PostgreSQL database. Please check your credentials.',
          duration: 5000
        })
      }
    } catch (err) {
      showErrorToast({
        title: 'Connection Error',
        description:
          err instanceof Error ? err.message : 'Failed to connect to PostgreSQL',
        duration: 5000
      })
    } finally {
      setIsConnecting(false)
    }
  }

  const updateField = (field: string, value: string | number) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const isFormValid = () => {
    const data = formData()
    return data.host && data.database && data.username && data.port > 0
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect to PostgreSQL</DialogTitle>
        </DialogHeader>

        <div class="grid gap-4 py-4">
          <TextField>
            <TextFieldLabel>Host</TextFieldLabel>
            <TextFieldInput
              type="text"
              value={formData().host}
              onInput={e => updateField('host', e.currentTarget.value)}
              placeholder="localhost"
            />
          </TextField>

          <TextField>
            <TextFieldLabel>Port</TextFieldLabel>
            <TextFieldInput
              type="number"
              value={formData().port}
              onInput={e => updateField('port', parseInt(e.currentTarget.value) || 5432)}
              placeholder="5432"
            />
          </TextField>

          <TextField>
            <TextFieldLabel>Database</TextFieldLabel>
            <TextFieldInput
              type="text"
              value={formData().database}
              onInput={e => updateField('database', e.currentTarget.value)}
              placeholder="database_name"
            />
          </TextField>

          <TextField>
            <TextFieldLabel>Username</TextFieldLabel>
            <TextFieldInput
              type="text"
              value={formData().username}
              onInput={e => updateField('username', e.currentTarget.value)}
              placeholder="username"
            />
          </TextField>

          <TextField>
            <TextFieldLabel>Password</TextFieldLabel>
            <TextFieldInput
              type="password"
              value={formData().password}
              onInput={e => updateField('password', e.currentTarget.value)}
              placeholder="password"
            />
          </TextField>

          <div class="flex flex-col gap-1">
            <label class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
              SSL Mode
            </label>
            <select
              value={formData().sslMode}
              onChange={e => updateField('sslMode', e.currentTarget.value)}
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="disable" selected>
                Disable
              </option>
              <option value="prefer">Prefer</option>
              <option value="require">Require</option>
            </select>
          </div>
        </div>

        <DialogFooter class="gap-2">
          <button
            onClick={() => props.onOpenChange(false)}
            class="px-4 py-2 text-sm bg-secondary text-secondary-foreground hover:bg-secondary/80 rounded-md transition-colors"
            disabled={isConnecting()}
          >
            Cancel
          </button>
          <button
            onClick={() => void handleConnect()}
            disabled={!isFormValid() || isConnecting()}
            class="px-4 py-2 bg-success  text-success-foreground text-sm rounded-md hover:bg-success/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Show when={isConnecting()} fallback="Connect">
              Connecting...
            </Show>
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default PostgresConnectionDialog

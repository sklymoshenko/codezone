import { Component, createSignal, Show } from 'solid-js'
import { DisconnectPostgreSQL } from 'wailsjs/go/main/App'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from './dialog'
import { showErrorToast } from './ErrorToast'

interface PostgresDisconnectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onDisconnect: (disconnected: boolean) => void
}

const PostgresDisconnectDialog: Component<PostgresDisconnectDialogProps> = props => {
  const [isDisconnecting, setIsDisconnecting] = createSignal(false)

  const handleDisconnect = async () => {
    setIsDisconnecting(true)

    try {
      await DisconnectPostgreSQL()

      // Update connection status
      props.onDisconnect(true)
      props.onOpenChange(false)
    } catch (err) {
      showErrorToast({
        title: 'Disconnection Error',
        description:
          err instanceof Error ? err.message : 'Failed to disconnect from PostgreSQL',
        duration: 5000
      })
    } finally {
      setIsDisconnecting(false)
    }
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Disconnect from PostgreSQL</DialogTitle>
        </DialogHeader>

        <div class="py-4">
          <p class="text-sm text-muted-foreground">
            Are you sure you want to disconnect from the PostgreSQL database?
          </p>
        </div>

        <DialogFooter class="gap-2">
          <button
            onClick={() => props.onOpenChange(false)}
            class="px-4 py-2 text-sm bg-secondary text-secondary-foreground hover:bg-secondary/80 rounded-md transition-colors"
            disabled={isDisconnecting()}
          >
            Cancel
          </button>
          <button
            onClick={() => void handleDisconnect()}
            disabled={isDisconnecting()}
            class="px-4 py-2 bg-destructive text-destructive-foreground text-sm rounded-md hover:bg-destructive/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <Show when={isDisconnecting()} fallback="Yes, Disconnect">
              Disconnecting...
            </Show>
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default PostgresDisconnectDialog

import { Language, PostgresConnectionStatus } from '~/types'
import { Show } from 'solid-js'
import type { executor } from 'wailsjs/go/models'
import SQLTable from './ui/SQLTable'

type OutputProps = {
  isExecuting: boolean
  executionResult: executor.ExecutionResult | null
  language: Language
  postgresConnectionStatus?: PostgresConnectionStatus
}

const Output = (props: OutputProps) => {
  const shouldShowTable = () => {
    return (
      props.language === 'postgres' &&
      props.executionResult?.sqlResult &&
      props.executionResult.sqlResult.queryType === 'SELECT' &&
      props.executionResult.sqlResult.rows &&
      props.executionResult.sqlResult.rows.length > 0
    )
  }

  const getTableRows = () => {
    if (!props.executionResult?.sqlResult?.rows) return []

    return props.executionResult.sqlResult.rows.map(row =>
      row.map(cell => {
        if (cell === null || cell === undefined) return '--'
        return String(cell)
      })
    )
  }

  return (
    <div class="w-full h-full flex flex-col bg-background">
      <div class="flex-1 overflow-auto font-mono text-base leading-normal">
        <Show when={shouldShowTable()}>
          <SQLTable
            columns={props.executionResult!.sqlResult!.columns}
            rows={getTableRows()}
            queryType={props.executionResult!.sqlResult!.queryType}
            executionTime={props.executionResult!.durationString}
            rowsCount={props.executionResult!.sqlResult!.rows.length}
          />
        </Show>

        <Show when={!shouldShowTable() && props.executionResult?.output}>
          <pre class="text-foreground whitespace-pre-wrap break-words m-0 p-4">
            {props.executionResult?.output}
          </pre>
        </Show>

        <Show when={props.executionResult?.error}>
          <pre class="text-destructive whitespace-pre-wrap break-words m-0 p-4">
            {props.executionResult?.error}
          </pre>
        </Show>
      </div>

      {/* Status bar */}
      <div class="flex-shrink-0 p-2 border-t bg-background/95 backdrop-blur text-xs text-muted-foreground tracking-wide font-mono flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Show
            when={props.isExecuting}
            fallback={
              <Show when={props.executionResult?.durationString}>
                <span>
                  Output{' '}
                  {props.executionResult?.durationString &&
                    ` (${props.executionResult.durationString})`}
                </span>
              </Show>
            }
          >
            <span class="italic">Running...</span>
          </Show>
        </div>

        <Show when={props.language === 'postgres'}>
          <div class="p-2 rounded-md border bg-card">
            <div class="flex items-center gap-2">
              <div
                class={`w-2 h-2 rounded-full`}
                classList={{
                  'animate-pulse bg-success':
                    props.postgresConnectionStatus === 'connected',
                  'bg-destructive': props.postgresConnectionStatus === 'disconnected'
                }}
              ></div>
              <span class="text-sm font-medium">
                PostgreSQL:{' '}
                {props.postgresConnectionStatus === 'connected'
                  ? 'Connected'
                  : 'Disconnected'}
              </span>
            </div>
          </div>
        </Show>
      </div>
    </div>
  )
}

export default Output

import { Language, PostgresConnectionStatus } from '~/types'
import { Show } from 'solid-js'
import type { executor } from 'wailsjs/go/models'

type OutputProps = {
  isExecuting: boolean
  executionResult: executor.ExecutionResult
  language: Language
  postgresConnectionStatus?: PostgresConnectionStatus
}

const Output = (props: OutputProps) => {
  return (
    <div class="relative flex-shrink-0 h-full p-4">
      <div class="absolute bottom-0 right-0 p-2 rounded-tl text-xs text-muted-foreground tracking-wide font-mono flex items-center gap-2">
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
      <div class="overflow-auto px-2 font-mono text-base leading-normal">
        <Show when={props.executionResult?.output}>
          <pre class="text-foreground whitespace-pre-wrap">
            {props.executionResult?.output}
          </pre>
        </Show>
        <Show when={props.executionResult?.error}>
          <pre class="text-destructive whitespace-pre-wrap">
            {props.executionResult?.error}
          </pre>
        </Show>
      </div>
    </div>
  )
}

export default Output

import { Show } from 'solid-js'
import type { executor } from 'wailsjs/go/models'

type OutputProps = {
  isExecuting: boolean
  executionResult: executor.ExecutionResult
}

const Output = (props: OutputProps) => {
  return (
    <div class="relative flex-shrink-0 h-full p-4">
      <div class="absolute bottom-0 right-0 p-2 rounded-tl text-xs text-muted-foreground tracking-wide font-mono">
        <Show
          when={props.isExecuting}
          fallback={
            <span>
              Output{' '}
              {props.executionResult?.durationString &&
                ` (${props.executionResult.durationString})`}
            </span>
          }
        >
          <span class="italic">Running...</span>
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

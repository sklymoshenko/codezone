import { Show } from 'solid-js'
import type { executor } from 'wailsjs/go/models'

type OutputProps = {
  isExecuting: boolean
  executionResult: executor.ExecutionResult
}

const Output = (props: OutputProps) => {
  return (
    <div class="relative flex-shrink-0 h-full">
      <div class="absolute bottom-0 right-0 p-2 rounded-tl text-xs text-gray-300 tracking-wide font-mono">
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
          <pre class="text-gray-200 whitespace-pre-wrap">
            {props.executionResult?.output}
          </pre>
        </Show>
        <Show when={props.executionResult?.error}>
          <pre class="text-red-400 whitespace-pre-wrap">
            {props.executionResult?.error}
          </pre>
        </Show>
      </div>
    </div>
  )
}

export default Output

import * as ToastPrimitive from '@kobalte/core/toast'
import { BiRegularErrorAlt } from 'solid-icons/bi'
import { BrowserOpenURL } from 'wailsjs/runtime/runtime'

interface ErrorToastProps {
  title: string
  description: string
  actionLabel?: string
  actionUrl?: string
  duration?: number
}

export function showErrorToast(props: ErrorToastProps) {
  ToastPrimitive.toaster.show(data => (
    <ToastPrimitive.Root
      toastId={data.toastId}
      duration={props.duration || 8000}
      class="group pointer-events-auto relative flex w-full max-w-md items-stretch overflow-hidden rounded-xl shadow-lg bg-card border"
    >
      {/* Left section - Dark theme content */}
      <div class="flex-1 bg-card p-2 flex items-start space-x-3">
        {/* Error Icon */}
        <BiRegularErrorAlt class="w-12 h-12 text-destructive" />

        {/* Content */}
        <div class="flex-1 min-w-0">
          <ToastPrimitive.Title class="text-sm font-semibold text-card-foreground">
            {props.title}
          </ToastPrimitive.Title>
          <ToastPrimitive.Description class="mt-1 text-xs text-muted-foreground">
            {props.description}
          </ToastPrimitive.Description>
          {props.actionLabel && props.actionUrl && (
            <button
              onClick={() => {
                BrowserOpenURL(props.actionUrl!)
              }}
              class="mt-2 inline-flex items-center text-xs font-medium text-primary hover:text-primary/80 transition-colors"
            >
              {props.actionLabel} →
            </button>
          )}
        </div>
      </div>

      {/* Close button */}
      <ToastPrimitive.CloseButton class="absolute right-2 top-2 text-muted-foreground hover:text-foreground transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </ToastPrimitive.CloseButton>
    </ToastPrimitive.Root>
  ))
}

export function debounce<F extends (...args: any[]) => void>(
  func: F,
  delay: number
): (...args: Parameters<F>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | undefined;

  // The returned function is an arrow function.
  return (...args: Parameters<F>) => {
    // Clear the previous timeout to reset the timer.
    clearTimeout(timeoutId);

    // Set a new timeout to call the function with spread arguments.
    timeoutId = setTimeout(() => func(...args), delay);
  };
} 
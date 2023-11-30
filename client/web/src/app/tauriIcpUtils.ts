/** Invoke a Tauri command */
export function tauriInvoke(command: string): void {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-explicit-any
    ;(window as any).__TAURI__?.invoke(command)
}

/**
 * Utility function to open the callback URL in Cody App. Used where
 * window.open or target="_blank" cannot be used.
 */
export function tauriShellOpen(uri: string): void {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-explicit-any
    ;(window as any).__TAURI__?.shell?.open(uri)
}

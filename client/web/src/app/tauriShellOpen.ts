/**
 * Utility function to open the callback URL in Sourcegraph App. Used where
 * window.open or target="_blank" cannot be used.
 */
export function tauriShellOpen(uri: string): void {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-explicit-any
    ;(window as any).__TAURI__?.shell?.open(uri)
}

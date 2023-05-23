export function tauriInvoke(command: string): void {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-explicit-any
    ;(window as any).__TAURI__?.invoke(command)
}

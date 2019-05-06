export function assertEnv(env: typeof window['EXTENSION_ENV']): void {
    if (window.EXTENSION_ENV !== env) {
        throw new Error(
            'Detected transitive import of an entrypoint! ' +
                window.EXTENSION_ENV +
                ' attempted to import a file that is only intended to be imported by ' +
                env +
                '.'
        )
    }
}

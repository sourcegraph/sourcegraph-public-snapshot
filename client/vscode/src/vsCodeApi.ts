declare global {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    // eslint-disable-next-line no-var
    var acquireVsCodeApi: <State = any>() => VsCodeApi<State>
}

/**
 * Built-in VS Code API exposed to webviews to communicate with the "Core" extension.
 * We typically use this as a low-level building block for the APIs used in our webviews
 * (wrapped w/ Comlink).
 */
export interface VsCodeApi<State = any> {
    postMessage: (message: any) => void
    getState: () => State | undefined
    setState: (state: State) => void
}

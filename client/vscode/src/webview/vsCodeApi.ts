declare global {
    // eslint-disable-next-line no-var
    var acquireVsCodeApi: <State = any>() => VsCodeApi<State>
}

export interface VsCodeApi<State = any> {
    postMessage: (message: any) => void
    getState: () => State | undefined
    setState: (state: State) => void
}

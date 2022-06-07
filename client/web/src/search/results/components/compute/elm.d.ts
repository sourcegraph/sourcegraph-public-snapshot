declare module '*.elm' {
    export const Elm: { Main: HTMLElement }
}

declare module 'react-elm-components'

interface ElmEvent {
    data: string
    eventType?: string
    id?: string
}

interface ExperimentalOptions {}

interface ComputeInput {
    computeQueries: string[]
    experimentalOptions: ExperimentalOptions
}

interface Ports {
    receiveEvent: { send: (event: ElmEvent) => void }
    openStream: { subscribe: (callback: (query: string) => void) => void }
    emitInput: { subscribe: (callback: (input: ComputeInput) => void) => void }
}

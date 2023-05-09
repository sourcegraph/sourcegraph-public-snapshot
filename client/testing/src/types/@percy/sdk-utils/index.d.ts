import logger2 from '@percy/logger'

declare module '@percy/sdk-utils' {
    interface PercyEnvironment {
        address: string
        enabled: boolean
        version: {
            major: number
            minor: number
            patch: number
            toString(): string
        }
        config: object
        domScript: string
    }

    interface SnapshotOptions {
        name: string
        url: string
        domSnapshot: string
        environmentInfo?: string[] | string
        clientInfo?: string
        widths?: number[]
        minHeight?: number
        enableJavaScript?: boolean
        requestHeaders?: object
    }

    interface RequestResponse {
        status: number
        statusText: string
        headers: any
        body: any
    }

    export const logger = logger2
    export const percy: PercyEnvironment
    export function isPercyEnabled(): Promise<boolean>
    export function fetchPercyDOM(): Promise<string>
    export function postSnapshot(options: SnapshotOptions): Promise<void>
    export function request(url: string, options?: object): Promise<RequestResponse>
    export namespace request {
        function fetch(url: string, options?: object): Promise<RequestResponse>
    }
}

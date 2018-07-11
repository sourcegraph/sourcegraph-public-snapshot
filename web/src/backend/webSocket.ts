import { Observable, Subject } from 'rxjs'
import { InitializeError, InitializeParams, InitializeResult } from 'vscode-languageserver/lib/main'
import * as rpc from 'vscode-ws-jsonrpc'
import { LSPContext, LSPRequest, ResponseResults } from './lsp'

const initialize = new rpc.RequestType<InitializeParams, InitializeResult, InitializeError, void>('initialize')

/** Reuse open connections (don't create a new WebSocket for each request). */
const CONNECTIONS = new Map<
    string,
    Promise<{
        conn: rpc.MessageConnection
        initializeResult: InitializeResult
    }>
>()

export function webSocketSendLSPRequest(ctx: LSPContext, request?: LSPRequest): Observable<ResponseResults> {
    const results = new Subject<ResponseResults>()

    // We include ?mode= in the url to make it easier to find the correct LSP
    // websocket connection in (e.g.) the Chrome network inspector. It does not
    // affect any behaviour.
    const url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/.api/lsp?mode=${
        ctx.mode
    }`
    const initializationOptions = { mode: ctx.mode, settings: ctx.settings }
    const key = `${url}|${JSON.stringify(ctx.settings)}`

    let conn = CONNECTIONS.get(key)
    if (!conn) {
        conn = new Promise((resolve, reject) => {
            rpc.listen({
                webSocket: new WebSocket(url),
                onConnection: (conn: rpc.MessageConnection) => {
                    conn.listen()
                    conn.onClose(() => CONNECTIONS.delete(url))
                    conn.onError(err => console.error(err))

                    const rootUri = `git://${ctx.repoPath}?${ctx.commitID}`
                    conn
                        .sendRequest(initialize, {
                            rootUri,
                            initializationOptions,
                        } as InitializeParams)
                        .then((initResult: InitializeResult) => {
                            resolve({ conn, initializeResult: initResult })
                        })
                },
            })
        })
        CONNECTIONS.set(key, conn)
    }

    conn.then(({ conn, initializeResult }) => {
        if (request) {
            conn.sendRequest<any>(request.method, request.params).then(result => {
                results.next([initializeResult, result, {}])
                results.complete()
            })
        } else {
            results.next([initializeResult, {}])
            results.complete()
        }
    })

    return results
}

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Sourcegraph. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
'use strict'

import * as vscode from 'vscode'
import * as path from 'path'
import {
    LanguageClient,
    RevealOutputChannelOn,
    LanguageClientOptions,
    ErrorCodes,
    MessageTransports,
    ProvideWorkspaceSymbolsSignature,
    ShowMessageParams,
    NotificationHandler,
    ResponseError,
    InitializeError,
} from 'vscode-languageclient'
import { v4 as uuidV4 } from 'uuid'
import { MessageTrace, webSocketStreamOpener } from './connection'
import * as log from './log'

function dispose(toDispose: vscode.Disposable[]): void {
    toDispose.forEach(disposable => disposable.dispose())
    toDispose.length = 0
}

const REUSE_BACKEND_LANG_SERVERS = true

/**
 * JSON-RPC 2.0 error codes returned by the LSP proxy.
 */
enum ProxyErrors {
    ModeNotFound = -32000,
}

/**
 * TODO(sqs): While we migrate from extractResourceInfo to
 * findContainingFolder/IFolderContainmentService, this function is exported and used by
 * both the workbench and extension host processes to determine the containing folder in a
 * limited but simple manner (that's equivalent to extractResourceInfo). In the future
 * this implementation will become more advanced to support repositories with varying
 * number of path components, etc.
 */
export function findContainingFolder(resource: vscode.Uri): vscode.Uri | undefined {
    if (resource.scheme === 'repo') {
        return resource.with({
            path: resource.path
                .split('/')
                .slice(0, 3)
                .join('/'),
        })
    }
    if (resource.scheme === 'repo+version') {
        return resource.with({
            path: resource.path
                .split('/')
                .slice(0, 3)
                .join('/'),
            // Preserve query component of URI.
        })
    }
    return undefined
}

export function toRelativePath(folder: vscode.Uri, resource: vscode.Uri): string | undefined {
    // Handle root with revision in querystring and resources with revision in
    // querystring.
    const folderString = folder.with({ query: '' }).toString(true /* skipEncoding */)
    const resourceString = resource.with({ query: '' }).toString(true /* skipEncoding */)

    const baseMatches = resourceString === folderString || resourceString.startsWith(folderString + '/')
    const queryMatches = (!folder.query && !resource.query) || folder.query === resource.query
    if (baseMatches && queryMatches) {
        return resourceString.slice(folderString.length + 1)
    }

    return undefined // resource is not inside folder
}

/**
 * Creates a new LSP client. The mode specifies which backend language
 * server to communicate with. The languageIds are the vscode document
 * languages that this client should be used to provide hovers, etc.,
 * for.
 */
export function newClient(
    mode: string,
    languageIds: string[],
    rootWithoutCommit: vscode.Uri,
    commitID: string
): LanguageClient {
    if (!commitID) {
        throw new Error(`no commit ID for workspace ${rootWithoutCommit}`)
    }

    const root = rootWithoutCommit.with({ query: commitID })

    const options: LanguageClientOptions = {
        revealOutputChannelOn: RevealOutputChannelOn.Never,
        documentSelector: languageIds.map(languageId => ({ language: languageId })),
        initializationOptions: {
            rootUri: root.toString(),
            mode: mode,
            session: REUSE_BACKEND_LANG_SERVERS ? undefined : uuidV4(),
        },
        initializationFailedHandler: (error: ResponseError<InitializeError> | Error | any): boolean => {
            if (error && error.code === ProxyErrors.ModeNotFound) {
                // Don't report to the user because we already show a nicer language
                // support warning.
                return false
            }

            if (error && error.message) {
                vscode.window.showErrorMessage(error.message)
            }
            return false
        },
        uriConverters: {
            code2Protocol: (value: vscode.Uri): string => {
                const gitDirPath = vscode.workspace.getWorkspaceFolder(value).uri.fsPath
                value = value.with({ path: value.path.slice(gitDirPath.length) })
                if (value.scheme === 'file') {
                    return value.toString()
                }
                if (value.scheme === 'repo' || value.scheme === 'repo+version') {
                    const folder = findContainingFolder(value)
                    if (folder) {
                        // Convert to the format that the LSP proxy server expects.
                        return vscode.Uri.parse(`git://${folder.authority}${folder.path}`)
                            .with({
                                query: commitID,
                                fragment: toRelativePath(folder, value),
                            })
                            .toString()
                    }
                }
                throw new Error(`unknown URI scheme in ${value.toString()}`)
            },
            protocol2Code: (value: string): vscode.Uri => {
                const uri = vscode.Uri.parse(value)
                if (uri.scheme === 'git') {
                    // URI is of the form git://github.com/owner/repo?gitrev#dir/file.

                    // Convert to repo://github.com/owner/repo/dir/file if in the same workspace.
                    if (uri.with({ scheme: 'repo', query: '', fragment: '' }).toString() === root.toString()) {
                        return root.with({
                            scheme: 'repo',
                            path: root.path + `${uri.fragment !== '' ? `/${decodeURIComponent(uri.fragment)}` : ''}`,
                        })
                    }

                    // Convert to repo+version://github.com/owner/repo/dir/file.txt?gitrev.
                    return uri.with({
                        scheme: 'repo+version',
                        path: uri.path.replace(/\/$/, '') + '/' + decodeURIComponent(uri.fragment),
                        fragment: '',
                    })
                }
                throw new Error('language server sent URI with unsupported scheme: ' + value)
            },
        },
    }

    const dummy = void 0 as any // dummy server arg (we override createMessageTransports to supply this)

    return new class WebSocketLanguageClient extends LanguageClient {
        // Override to use a WebSocket transport instead of a StreamInfo (which requires a
        // duplex stream).
        protected createMessageTransports(encoding: string): Thenable<MessageTransports> {
            const endpoint = vscode.Uri.parse(vscode.workspace.getConfiguration('sourcegraph').get<string>('URL'))
            // const wsOrigin = endpoint.with({ scheme: endpoint.scheme === 'http' ? 'ws' : 'wss' });

            // We include ?mode= in the url to make it easier to find the correct LSP
            // websocket connection in (e.g.) the Chrome network inspector. It does not
            // affect any behaviour.
            const url = `ws://${endpoint.authority}/.api/lsp`
            return webSocketStreamOpener(url, createRequestTracer(mode))
        }
    }('lsp-' + mode, 'lsp-' + mode, dummy, options)
}

let traceOutputChannel: vscode.OutputChannel | undefined

function createRequestTracer(languageId: string): ((trace: MessageTrace) => void) | undefined {
    return (trace: MessageTrace) => {
        if (!vscode.workspace.getConfiguration('lsp').get<boolean>('trace')) {
            return undefined
        }

        if (!traceOutputChannel) {
            traceOutputChannel = vscode.window.createOutputChannel('LSP (trace)')
        }

        let label: string
        if (!trace.response.error) {
            label = 'OK '
        } else if (trace.response.error.code === ErrorCodes.RequestCancelled) {
            label = 'CXL'
        } else {
            label = 'ERR'
        }
        traceOutputChannel.appendLine(
            `${label} ${languageId} ${describeRequest(trace.request.method, trace.request.params)} — ${trace.endTime -
                trace.startTime}ms`
        )
        if (trace.response.meta && trace.response.meta['X-Trace']) {
            traceOutputChannel.appendLine(` - Trace: ${trace.response.meta['X-Trace']}`)
        }
        traceOutputChannel.appendLine('')
        console.log('Request Params:', trace.request.params)
        console.log('Response:', trace.response)
    }
}

function describeRequest(method: string, params: any): string {
    if (params.textDocument && params.textDocument.uri && params.position) {
        return `${method} @ ${params.position.line + 1}:${params.position.character + 1}`
    }
    if (typeof params.query !== 'undefined') {
        return `${method} with query ${JSON.stringify(params.query)}`
    }
    if (params.rootPath) {
        return `${method} ${params.rootPath}`
    }
    return method
}

import * as vscode from 'vscode'
import { LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/prompt/prompt-mixin'

import { start } from './main'

import { getConfiguration } from './configuration'
import { ExtensionApi } from './extension-api'
import { CODY_ACCESS_TOKEN_SECRET, VSCodeSecretStorage } from './secret-storage'

let client: LanguageClient

function sendCommandRequest(command: string, args: any[] | undefined): void {
    client
        .sendRequest('workspace/executeCommand', {
            command,
            arguments: args,
        })
        .catch(e => {
            console.error(e)
        })
}

export function activate(context: vscode.ExtensionContext): ExtensionApi {
    console.log('Cody extension activated')

    let serverOptions: ServerOptions = {
        run: {
            command: '/home/pjlast/go/bin/llmsp',
            transport: TransportKind.stdio,
        },
        debug: {
            command: '/home/pjlast/go/bin/llmsp',
            transport: TransportKind.stdio,
        },
    }

    PromptMixin.add(languagePromptMixin(vscode.env.language))

    vscode.commands.registerCommand

    // Options to control the language client
    let clientOptions: LanguageClientOptions = {
        // Register the server for plain text documents
        documentSelector: [{ scheme: 'file', language: 'go' }],
        synchronize: {
            // Notify the server about file changes to '.clientrc files contained in the workspace
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.go'),
        },
        middleware: {
            resolveCodeAction: async (item, token, next): Promise<vscode.CodeAction | undefined> => {
                const action = await next(item, token)
                if (action != null && action != undefined && action.command != undefined) {
                    sendCommandRequest(action.command.command, action.command.arguments)
                }
                return undefined
            },
        },
    }

    // Create the language client and start the client.
    client = new LanguageClient('llmsp', 'LLM-powered LSP', serverOptions, clientOptions)
    const secretStorage = new VSCodeSecretStorage(context.secrets)

    // Start the client. This will also launch the server
    client.start().catch(e => {
        console.error("LSP failed to start: ", e)
    })

    const config = getConfiguration(vscode.workspace.getConfiguration())

    secretStorage.get(CODY_ACCESS_TOKEN_SECRET).then(res => {
        client.sendNotification('workspace/didChangeConfiguration', {
            settings: {
                llmsp: {
                    sourcegraph: {
                        url: config.serverEndpoint,
                        accessToken: res ?? '',
                    },
                },
            },
        })
    })

    PromptMixin.add(languagePromptMixin(vscode.env.language))

    if (process.env.CODY_FOCUS_ON_STARTUP) {
        setTimeout(() => {
            void vscode.commands.executeCommand('cody.chat.focus')
        }, 250)
    }

    start(context)
        .then(disposable => context.subscriptions.push(disposable))
        .catch(error => console.error(error))

    return new ExtensionApi()
}

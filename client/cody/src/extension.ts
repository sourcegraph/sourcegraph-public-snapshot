import * as os from 'os'
import path from 'path'

import * as vscode from 'vscode'
import { LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/prompt/prompt-mixin'

import { start } from './main'

import { getConfiguration } from './configuration'
import { ExtensionApi } from './extension-api'
import { CODY_ACCESS_TOKEN_SECRET, VSCodeSecretStorage } from './secret-storage'

let client: LanguageClient

export async function sendCommandRequest(command: string, args: any[] | undefined): Promise<void> {
    return client
        .sendRequest('workspace/executeCommand', {
            command,
            arguments: args,
        })
}

export function activate(context: vscode.ExtensionContext): ExtensionApi {
    console.log('Cody extension activated')

    const arch = process.env.npm_config_path || os.arch()
    let binaryName = 'llmsp-v1.0.0'
    switch (arch) {
        case 'arm64':
            binaryName += '-arm64'
            break
        case 'amd64':
            binaryName += '-amd64'
            break
    }

    switch (os.platform()) {
        case 'darwin':
            binaryName += '-darwin'
            break
        case 'win32':
            binaryName += '-windows'
            break
        case 'linux':
            binaryName += '-linux'
            break
    }

    let serverOptions: ServerOptions = {
        run: {
            command: path.join(context.extensionPath, 'resources', 'bin', binaryName),
            transport: TransportKind.stdio,
        },
        debug: {
            command: path.join(context.extensionPath, 'resources', 'bin', binaryName),
            transport: TransportKind.stdio,
        },
    }

    PromptMixin.add(languagePromptMixin(vscode.env.language))

    // Options to control the language client
    let clientOptions: LanguageClientOptions = {
        // Register the server for plain text documents
        documentSelector: [{ scheme: 'file', language: 'go' }],
        synchronize: {
            // Notify the server about file changes to '.clientrc files contained in the workspace
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.go'),
        },
        middleware: {
            resolveCodeAction: async (item, token, next): Promise<vscode.CodeAction | null | undefined> => {
                const action = await next(item, token)
                if (action != null && action != undefined && action.command != undefined) {
                    try {
                        await sendCommandRequest(action.command.command, action.command.arguments)
                    } catch (err) {
                        console.error(err)
                    }
                }
                return action
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
    const repos = config.codebase != undefined && config.codebase != '' ? [config.codebase] : null

    context.subscriptions.push(
        vscode.commands.registerCommand('cody.dostuff', async (args: any[]) => {
            const editor = vscode.window.activeTextEditor;
            if (editor) {
                const filePath = "file://" + editor.document.uri.fsPath;
                const selection = editor.selection;
                const start = selection.start;
                const end = selection.end;
                const prompt = args?.length ? (args[0] as string) : await vscode.window.showInputBox()
                try {
                    await sendCommandRequest("cody", [filePath, start.line, end.line, prompt, true, true])
                } catch (err) {
                    console.error(err)
                }
            }
        }))

    secretStorage.get(CODY_ACCESS_TOKEN_SECRET).then(res => {
        client.sendNotification('workspace/didChangeConfiguration', {
            settings: {
                llmsp: {
                    sourcegraph: {
                        url: config.serverEndpoint,
                        accessToken: res ?? '',
                        repos: repos,
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

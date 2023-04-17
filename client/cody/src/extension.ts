import * as os from 'os'
import path from 'path'

import * as vscode from 'vscode'
import { LanguageClient, LanguageClientOptions, ServerOptions, TransportKind } from 'vscode-languageclient/node'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/prompt/prompt-mixin'

import { start } from './main'

import { getConfiguration } from './configuration'
import { ExtensionApi } from './extension-api'
import { CODY_ACCESS_TOKEN_SECRET, VSCodeSecretStorage } from './secret-storage'

// This client can be used to execute arbitrary commands.
let client: LanguageClient

export async function sendCommandRequest<R>(command: string, args: any[] | undefined): Promise<R> {
    return client
        .sendRequest<R>('workspace/executeCommand', {
            command,
            arguments: args,
        })
}

type ErrorAnswer = {
    answer: string
}

function getLLMSPBinary(): string {
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

    return binaryName
}

export async function activate(context: vscode.ExtensionContext): Promise<ExtensionApi> {
    console.log('Cody extension activated')

    const llmspBinary = getLLMSPBinary()

    const serverOptions: ServerOptions = {
        run: {
            command: path.join(context.extensionPath, 'resources', 'bin', llmspBinary),
            transport: TransportKind.stdio,
        },
        debug: {
            command: path.join(context.extensionPath, 'resources', 'bin', llmspBinary),
            transport: TransportKind.stdio,
        },
    }

    PromptMixin.add(languagePromptMixin(vscode.env.language))

    // Options to control the language client
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'go' }, { scheme: 'file', language: 'typescript' }], // TODO: Support more (or all) languages
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('*'),
        },
        middleware: {
            resolveCodeAction: async (item, token, next): Promise<vscode.CodeAction | null | undefined> => {
                const action = await next(item, token)
                if (action != null && action != undefined && action.command != undefined) {
                    // We can intercept certain code actions and handle special cases here
                    if (action.command.command === "cody.explainErrors") {
                        try {
                            let resp = await sendCommandRequest<ErrorAnswer>(action.command.command, action.command.arguments)
                            // Display response in chat window
                            return action
                        } catch (err) {
                        }
                    }
                    try {
                        await sendCommandRequest<void>(action.command.command, action.command.arguments)
                    } catch (err) {
                        console.error(err)
                    }
                }
                return action
            },
        },
    }

    // Create the language client.
    client = new LanguageClient('llmsp', 'LLM-powered LSP', serverOptions, clientOptions)
    const secretStorage = new VSCodeSecretStorage(context.secrets)

    // Start the client. This will also launch the server
    client.start().catch(e => {
        console.error("LSP failed to start: ", e)
    })

    const config = getConfiguration(vscode.workspace.getConfiguration())
    const repos = config.codebase != undefined && config.codebase != '' ? [config.codebase] : null

    // Registers a command to instruct Cody to do something with the selected text.
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

    const codyAccessToken = await secretStorage.get(CODY_ACCESS_TOKEN_SECRET)

    client.sendNotification('workspace/didChangeConfiguration', {
        settings: {
            llmsp: {
                sourcegraph: {
                    url: config.serverEndpoint,
                    accessToken: codyAccessToken ?? '',
                    autoComplete: 'always',
                    repos: repos,
                },
            },
        },
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

import * as vscode from 'vscode'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { CodyCompletionItemProvider } from '../completions'
import { CompletionsDocumentProvider } from '../completions/docprovider'
import { History } from '../completions/history'
import { getConfiguration } from '../configuration'
import { VSCodeEditor } from '../editor/vscode-editor'
import { logEvent, updateEventLogger } from '../event-logger'
import { ExtensionApi } from '../extension-api'
import { configureExternalServices } from '../external-services'
import { getRgPath } from '../rg'
import { sanitizeCodebase, sanitizeServerEndpoint } from '../sanitize'
import { CODY_ACCESS_TOKEN_SECRET, InMemorySecretStorage, SecretStorage, VSCodeSecretStorage } from '../secret-storage'

import { LocalStorage } from './LocalStorageProvider'

function getSecretStorage(context: vscode.ExtensionContext): SecretStorage {
    return process.env.CODY_TESTING === 'true' ? new InMemorySecretStorage() : new VSCodeSecretStorage(context.secrets)
}

// Registers Commands and Webview at extension start up
export const CommandsProvider = async (context: vscode.ExtensionContext): Promise<ExtensionApi> => {
    // for tests
    const extensionApi = new ExtensionApi()

    const secretStorage = getSecretStorage(context)
    const localStorage = new LocalStorage(context.globalState)
    const config = getConfiguration(vscode.workspace.getConfiguration())

    await updateEventLogger(config, secretStorage, localStorage)

    const editor = new VSCodeEditor()
    const rgPath = await getRgPath(context.extensionPath)
    const mode = config.debug ? 'development' : 'production'
    const serverEndpoint = sanitizeServerEndpoint(config.serverEndpoint)
    const codebase = sanitizeCodebase(config.codebase)

    const { intentDetector, codebaseContext, chatClient, completionsClient } = await configureExternalServices(
        serverEndpoint,
        codebase,
        rgPath,
        editor,
        secretStorage,
        config.useContext,
        mode,
        config.customHeaders
    )

    // Create chat webview
    const chatProvider = ChatViewProvider.create(
        context.extensionPath,
        sanitizeCodebase(config.codebase),
        sanitizeServerEndpoint(config.serverEndpoint),
        config.useContext,
        secretStorage,
        localStorage,
        editor,
        rgPath,
        mode,
        intentDetector,
        codebaseContext,
        chatClient,
        config.customHeaders
    )

    vscode.window.registerWebviewViewProvider('cody.chat', chatProvider, {
        webviewOptions: { retainContextWhenHidden: true },
    })

    await vscode.commands.executeCommand('setContext', 'cody.activated', true)

    const disposables: vscode.Disposable[] = []

    disposables.push(
        // Toggle Chat
        vscode.commands.registerCommand('cody.toggle-enabled', async () => {
            const workspaceConfig = vscode.workspace.getConfiguration()
            const config = getConfiguration(workspaceConfig)

            await workspaceConfig.update(
                'cody.enabled',
                !workspaceConfig.get('cody.enabled'),
                vscode.ConfigurationTarget.Global
            )
            logEvent(
                'CodyVSCodeExtension:codyToggleEnabled:clicked',
                { serverEndpoint: config.serverEndpoint },
                { serverEndpoint: config.serverEndpoint }
            )
        }),
        // Access token
        vscode.commands.registerCommand('cody.set-access-token', async (args: any[]) => {
            const config = getConfiguration(vscode.workspace.getConfiguration())
            const tokenInput = args?.length ? (args[0] as string) : await vscode.window.showInputBox()
            if (tokenInput === undefined || tokenInput === '') {
                return
            }
            await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, tokenInput)
            logEvent(
                'CodyVSCodeExtension:codySetAccessToken:clicked',
                { serverEndpoint: config.serverEndpoint },
                { serverEndpoint: config.serverEndpoint }
            )
        }),
        vscode.commands.registerCommand('cody.delete-access-token', async () => {
            const config = getConfiguration(vscode.workspace.getConfiguration())
            await secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
            logEvent(
                'CodyVSCodeExtension:codyDeleteAccessToken:clicked',
                { serverEndpoint: config.serverEndpoint },
                { serverEndpoint: config.serverEndpoint }
            )
        }),
        // TOS
        vscode.commands.registerCommand('cody.accept-tos', version =>
            localStorage.set('cody.tos-version-accepted', version)
        ),
        vscode.commands.registerCommand('cody.get-accepted-tos-version', () =>
            localStorage.get('cody.tos-version-accepted')
        ),
        // Commands
        vscode.commands.registerCommand('cody.recipe.explain-code', async () => executeRecipe('explain-code-detailed')),
        vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () =>
            executeRecipe('explain-code-high-level')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () =>
            executeRecipe('generate-unit-test')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-docstring', async () =>
            executeRecipe('generate-docstring')
        ),
        vscode.commands.registerCommand('cody.recipe.translate-to-language', async () =>
            executeRecipe('translate-to-language')
        ),
        vscode.commands.registerCommand('cody.recipe.git-history', async () => executeRecipe('git-history')),
        vscode.commands.registerCommand('cody.recipe.improve-variable-names', async () =>
            executeRecipe('improve-variable-names')
        )
    )

    if (config.experimentalSuggest) {
        const docprovider = new CompletionsDocumentProvider()
        vscode.workspace.registerTextDocumentContentProvider('cody', docprovider)

        const history = new History()
        const completionsProvider = new CodyCompletionItemProvider(completionsClient, docprovider, history)
        context.subscriptions.push(
            vscode.commands.registerCommand('cody.experimental.suggest', async () => {
                await completionsProvider.fetchAndShowCompletions()
            })
        )
        context.subscriptions.push(
            vscode.languages.registerInlineCompletionItemProvider({ scheme: 'file' }, completionsProvider)
        )
    }

    // Watch all relevant configuration and secrets for changes.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(async event => {
            if (event.affectsConfiguration('cody') || event.affectsConfiguration('sourcegraph')) {
                const config = getConfiguration(vscode.workspace.getConfiguration())
                await updateEventLogger(config, secretStorage, localStorage)
                await chatProvider.onConfigChange(
                    'endpoint',
                    sanitizeCodebase(config.codebase),
                    sanitizeServerEndpoint(config.serverEndpoint)
                )
            }
        })
    )

    context.subscriptions.push(
        secretStorage.onDidChange(async key => {
            if (key === CODY_ACCESS_TOKEN_SECRET) {
                const config = getConfiguration(vscode.workspace.getConfiguration())
                await updateEventLogger(config, secretStorage, localStorage)
                await chatProvider
                    .onConfigChange(
                        'token',
                        sanitizeCodebase(config.codebase),
                        sanitizeServerEndpoint(config.serverEndpoint)
                    )
                    .catch(error => console.error(error))
            }
        })
    )

    const executeRecipe = async (recipe: string): Promise<void> => {
        await vscode.commands.executeCommand('cody.chat.focus')
        await chatProvider.executeRecipe(recipe)
    }

    vscode.Disposable.from(...disposables)

    return extensionApi
}

import * as openai from 'openai'
import * as vscode from 'vscode'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { CodyCompletionItemProvider } from '../completions'
import { CompletionsDocumentProvider } from '../completions/docprovider'
import { getConfiguration } from '../configuration'
import { ExtensionApi } from '../extension-api'

import { EventLogger } from './eventLogger'
import { LocalStorage } from './LocalStorageProvider'
import { CODY_ACCESS_TOKEN_SECRET, InMemorySecretStorage, SecretStorage, VSCodeSecretStorage } from './secret-storage'

function getSecretStorage(context: vscode.ExtensionContext): SecretStorage {
    return process.env.CODY_TESTING === 'true' ? new InMemorySecretStorage() : new VSCodeSecretStorage(context.secrets)
}

function sanitizeCodebase(codebase: string | undefined): string {
    if (!codebase) {
        return ''
    }
    const protocolRegexp = /^(https?):\/\//
    return codebase.replace(protocolRegexp, '')
}

function sanitizeServerEndpoint(serverEndpoint: string): string {
    const trailingSlashRegexp = /\/$/
    return serverEndpoint.trim().replace(trailingSlashRegexp, '')
}

// Registers Commands and Webview at extension start up
export const CommandsProvider = async (context: vscode.ExtensionContext): Promise<ExtensionApi> => {
    // for tests
    const extensionApi = new ExtensionApi()

    const secretStorage = getSecretStorage(context)
    const localStorage = new LocalStorage(context.globalState)
    const config = getConfiguration(vscode.workspace.getConfiguration())

    // Create chat webview
    const chatProvider = await ChatViewProvider.create(
        context.extensionPath,
        sanitizeCodebase(config.codebase),
        sanitizeServerEndpoint(config.serverEndpoint),
        config.useContext,
        config.debug,
        secretStorage,
        localStorage
    )

    vscode.window.registerWebviewViewProvider('cody.chat', chatProvider, {
        webviewOptions: { retainContextWhenHidden: true },
    })

    await vscode.commands.executeCommand('setContext', 'cody.activated', true)

    const disposables: vscode.Disposable[] = []

    disposables.push(
        // Toggle Chat
        vscode.commands.registerCommand('cody.toggle-enabled', async () => {
            const config = vscode.workspace.getConfiguration()
            await config.update('cody.enabled', !config.get('cody.enabled'), vscode.ConfigurationTarget.Global)
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyToggleEnabled:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        // Access token
        vscode.commands.registerCommand('cody.set-access-token', async (args: any[]) => {
            const tokenInput = args?.length ? (args[0] as string) : await vscode.window.showInputBox()
            if (tokenInput === undefined || tokenInput === '') {
                return
            }
            await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, tokenInput)
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codySetAccessToken:clicked')
                if (tokenInput) {
                    EventLogger.log('CodyVSCodeExtension:codySetAccessToken:clicked:tokenSet')
                } else {
                    EventLogger.log('CodyVSCodeExtension:codySetAccessToken:clicked:noTokenSet')
                }
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.delete-access-token', async () => {
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyDeleteAccessToken:clicked')
                if (!secretStorage.get(CODY_ACCESS_TOKEN_SECRET)) {
                    EventLogger.log('CodyVSCodeExtension:codyDeleteAccessToken:clicked:noToken')
                } else {
                    EventLogger.log('CodyVSCodeExtension:codyDeleteAccessToken:clicked:tokenExists')
                }
            } catch (error) {
                console.log(error)
            }
            secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
        }),
        // TOS
        vscode.commands.registerCommand('cody.accept-tos', version =>
            localStorage.set('cody.tos-version-accepted', version)
        ),
        vscode.commands.registerCommand('cody.get-accepted-tos-version', () =>
            localStorage.get('cody.tos-version-accepted')
        ),
        // Commands
        vscode.commands.registerCommand('cody.recipe.explain-code', async () => {
            executeRecipe('explain-code-detailed')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyExplainCode:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () => {
            executeRecipe('explain-code-high-level')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyExplainCodeHighLevel:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () => {
            executeRecipe('generate-unit-test')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyGenerateUnitTest:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.recipe.generate-docstring', async () => {
            executeRecipe('generate-docstring')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyGenerateDocstring:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.recipe.translate-to-language', async () => {
            executeRecipe('translate-to-language')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyTranslateToLanguage:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.recipe.git-history', async () => {
            executeRecipe('git-history')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyGitHistory:clicked')
            } catch (error) {
                console.log(error)
            }
        }),
        vscode.commands.registerCommand('cody.recipe.improve-variable-names', async () => {
            executeRecipe('improve-variable-names')
            // log event
            try {
                EventLogger.log('CodyVSCodeExtension:codyImproveVariableNames:clicked')
            } catch (error) {
                console.log(error)
            }
        })
    )

    if (config.experimentalSuggest && config.openaiKey) {
        const configuration = new openai.Configuration({
            apiKey: config.openaiKey,
        })
        const openaiApi = new openai.OpenAIApi(configuration)
        const docprovider = new CompletionsDocumentProvider()
        vscode.workspace.registerTextDocumentContentProvider('cody', docprovider)

        const completionsProvider = new CodyCompletionItemProvider(openaiApi, docprovider)
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
                await chatProvider.onConfigChange(
                    'endpoint',
                    sanitizeCodebase(config.codebase),
                    sanitizeServerEndpoint(config.serverEndpoint)
                )
            }
        })
    )

    context.subscriptions.push(
        secretStorage.onDidChange(key => {
            if (key === CODY_ACCESS_TOKEN_SECRET) {
                const config = getConfiguration(vscode.workspace.getConfiguration())
                chatProvider
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

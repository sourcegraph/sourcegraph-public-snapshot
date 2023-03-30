import * as openai from 'openai'
import * as vscode from 'vscode'

import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { EventLogger } from '@sourcegraph/cody-shared/src/telemetry/EventLogger'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { CodyCompletionItemProvider } from '../completions'
import { CompletionsDocumentProvider } from '../completions/docprovider'
import { getConfiguration } from '../configuration'
import { ExtensionApi } from '../extension-api'

import { LocalStorage } from './LocalStorageProvider'
import {
    CODY_ACCESS_TOKEN_SECRET,
    getAccessToken,
    InMemorySecretStorage,
    SecretStorage,
    VSCodeSecretStorage,
} from './secret-storage'

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

    const accessToken = await getAccessToken(secretStorage)
    const gqlAPIClient = new SourcegraphGraphQLAPIClient(sanitizeServerEndpoint(config.serverEndpoint), accessToken)
    const eventLogger = new EventLogger(localStorage, gqlAPIClient)

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
            await eventLogger.log('CodyVSCodeExtension:codyToggleEnabled:clicked')
        }),
        // Access token
        vscode.commands.registerCommand('cody.set-access-token', async (args: any[]) => {
            const tokenInput = args?.length ? (args[0] as string) : await vscode.window.showInputBox()
            if (tokenInput === undefined || tokenInput === '') {
                return
            }
            await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, tokenInput)
            await eventLogger.log('CodyVSCodeExtension:codySetAccessToken:clicked')
        }),
        vscode.commands.registerCommand('cody.delete-access-token', async () => {
            secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
            eventLogger.log('CodyVSCodeExtension:codyDeleteAccessToken:clicked')
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
            await eventLogger.log('CodyVSCodeExtension:askCodyExplainCode:clicked')
            executeRecipe('explain-code-detailed')
        }),
        vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () => {
            executeRecipe('explain-code-high-level')
            await eventLogger.log('CodyVSCodeExtension:codyExplainCodeHighLevel:clicked')
        }),
        vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () => {
            executeRecipe('generate-unit-test')
            await eventLogger.log('CodyVSCodeExtension:codyGenerateUnitTest:clicked')
        }),
        vscode.commands.registerCommand('cody.recipe.generate-docstring', async () => {
            executeRecipe('generate-docstring')
            await eventLogger.log('CodyVSCodeExtension:codyGenerateDocstring:clicked')
        }),
        vscode.commands.registerCommand('cody.recipe.translate-to-language', async () => {
            executeRecipe('translate-to-language')
            await eventLogger.log('CodyVSCodeExtension:codyTranslateToLanguage:clicked')
        }),
        vscode.commands.registerCommand('cody.recipe.git-history', async () => {
            executeRecipe('git-history')
            await eventLogger.log('CodyVSCodeExtension:codyGitHistory:clicked')
        }),
        vscode.commands.registerCommand('cody.recipe.improve-variable-names', async () => {
            executeRecipe('improve-variable-names')
            await eventLogger.log('CodyVSCodeExtension:codyImproveVariableNames:clicked')
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

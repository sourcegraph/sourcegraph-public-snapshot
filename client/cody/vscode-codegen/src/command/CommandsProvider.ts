import * as openai from 'openai'
import * as vscode from 'vscode'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { CodyCompletionItemProvider } from '../completions'
import { CompletionsDocumentProvider } from '../completions/docprovider'
import { CODY_ACCESS_TOKEN_SECRET, ConfigurationUseContext, getConfiguration } from '../configuration'
import { ExtensionApi } from '../extension-api'

import { InMemorySecretStorage, SecretStorage, VSCodeSecretStorage } from './secret-storage'

function getSecretStorage(context: vscode.ExtensionContext): SecretStorage {
    return process.env.CODY_TESTING === 'true' ? new InMemorySecretStorage() : new VSCodeSecretStorage(context.secrets)
}

// Registers Commands and Webview at extension start up
export const CommandsProvider = async (context: vscode.ExtensionContext): Promise<ExtensionApi> => {
    // for tests
    const extensionApi = new ExtensionApi()

    const secretStorage = getSecretStorage(context)
    const config = getConfiguration(vscode.workspace.getConfiguration())
    const accessToken = (await secretStorage.get(CODY_ACCESS_TOKEN_SECRET)) || ''
    const useContext: ConfigurationUseContext = config.useContext

    // Create chat webview
    const chatProvider = new ChatViewProvider(
        config.codebase || '',
        context.extensionPath,
        config.serverEndpoint,
        accessToken,
        config.embeddingsEndpoint,
        useContext,
        config.debug,
        secretStorage
    )

    vscode.window.registerWebviewViewProvider('cody.chat', chatProvider)

    await vscode.commands.executeCommand('setContext', 'sourcegraph.cody.activated', true)

    const disposables: vscode.Disposable[] = []

    disposables.push(
        // Toggle Chat
        vscode.commands.registerCommand('sourcegraph.cody.toggleEnabled', async () => {
            const config = vscode.workspace.getConfiguration()
            await config.update(
                'sourcegraph.cody.enable',
                !config.get('sourcegraph.cody.enable'),
                vscode.ConfigurationTarget.Global
            )
        }),
        // Access token
        vscode.commands.registerCommand('cody.set-access-token', async (args: any[]) => {
            const tokenInput = args?.length ? (args[0] as string) : await vscode.window.showInputBox()
            if (tokenInput === undefined || tokenInput === '') {
                return
            }
            await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, tokenInput)
        }),
        vscode.commands.registerCommand('cody.delete-access-token', async () =>
            secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
        ),
        // TOS
        vscode.commands.registerCommand('cody.accept-tos', async version => {
            if (typeof version !== 'number') {
                // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
                void vscode.window.showErrorMessage(`TOS version was not a number: ${version}`)
                return
            }
            await context.globalState.update('cody.tos-version-accepted', version)
        }),
        vscode.commands.registerCommand('cody.get-accepted-tos-version', async () => {
            const version = await context.globalState.get('cody.tos-version-accepted')
            return version
        }),
        // Commands
        vscode.commands.registerCommand('cody.recipe.explain-code', async () => executeRecipe('explainCode')),
        vscode.commands.registerCommand('cody.recipe.explain-code-high-level', async () =>
            executeRecipe('explainCodeHighLevel')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-unit-test', async () =>
            executeRecipe('generateUnitTest')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-docstring', async () =>
            executeRecipe('generateDocstring')
        ),
        vscode.commands.registerCommand('cody.recipe.translate-to-language', async () =>
            executeRecipe('translateToLanguage')
        ),
        vscode.commands.registerCommand('cody.recipe.git-history', async () => executeRecipe('gitHistory'))
    )

    if (config.experimentalSuggest) {
        const openaiKey = vscode.workspace.getConfiguration().get<string>('cody.keys.openai')
        const configuration = new openai.Configuration({
            apiKey: openaiKey,
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
                await chatProvider.configChangeDetected('endpoints')
            }
        })
    )
    context.subscriptions.push(
        secretStorage.onDidChange(async key => {
            if (key === CODY_ACCESS_TOKEN_SECRET) {
                await chatProvider.configChangeDetected('token')
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

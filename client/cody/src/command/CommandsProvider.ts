import * as vscode from 'vscode'

import { ChatViewProvider } from '../chat/ChatViewProvider'
import { getConfiguration } from '../configuration'
import { ExtensionApi } from '../extension-api'

import { LocalStorage } from './LocalStorageProvider'
import { CODY_ACCESS_TOKEN_SECRET, InMemorySecretStorage, SecretStorage, VSCodeSecretStorage } from './secret-storage'

function getSecretStorage(context: vscode.ExtensionContext): SecretStorage {
    return process.env.CODY_TESTING === 'true' ? new InMemorySecretStorage() : new VSCodeSecretStorage(context.secrets)
}

// Registers Commands and Webview at extension start up
export const CommandsProvider = async (context: vscode.ExtensionContext): Promise<ExtensionApi> => {
    // for tests
    const extensionApi = new ExtensionApi()

    const secretStorage = getSecretStorage(context)
    const config = getConfiguration(vscode.workspace.getConfiguration())
    const localStorage = new LocalStorage(context.globalState)

    // Create chat webview
    const chatProvider = await ChatViewProvider.create(
        context.extensionPath,
        config.codebase ?? '',
        config.serverEndpoint,
        config.useContext,
        config.debug,
        secretStorage,
        localStorage
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
        vscode.commands.registerCommand(
            'cody.accept-tos',
            async version => await localStorage.set('cody.tos-version-accepted', version)
        ),
        vscode.commands.registerCommand(
            'cody.get-accepted-tos-version',
            async () => await localStorage.get('cody.tos-version-accepted')
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
        vscode.commands.registerCommand('cody.recipe.git-history', async () => executeRecipe('git-history'))
    )

    // Watch all relevant configuration and secrets for changes.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(async event => {
            if (event.affectsConfiguration('cody') || event.affectsConfiguration('sourcegraph')) {
                const config = getConfiguration(vscode.workspace.getConfiguration())
                await chatProvider.onConfigChange('endpoint', config.codebase ?? '', config.serverEndpoint)
            }
        })
    )
    context.subscriptions.push(
        secretStorage.onDidChange(async key => {
            if (key === CODY_ACCESS_TOKEN_SECRET) {
                const config = getConfiguration(vscode.workspace.getConfiguration())
                await chatProvider.onConfigChange('token', config.codebase ?? '', config.serverEndpoint)
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

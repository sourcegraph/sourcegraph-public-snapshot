import * as vscode from 'vscode'

import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'

import { ChatViewProvider, getAuthStatus } from './chat/ChatViewProvider'
import { DOTCOM_URL, LOCAL_APP_URL, isLoggedIn } from './chat/protocol'
import { CodyCompletionItemProvider } from './completions'
import { CompletionsDocumentProvider } from './completions/docprovider'
import { History } from './completions/history'
import * as CompletionsLogger from './completions/logger'
import { getConfiguration, getFullConfig } from './configuration'
import { VSCodeEditor } from './editor/vscode-editor'
import { logEvent, updateEventLogger } from './event-logger'
import { configureExternalServices } from './external-services'
import { FixupController } from './non-stop/FixupController'
import { getRgPath } from './rg'
import { GuardrailsProvider } from './services/GuardrailsProvider'
import { InlineController } from './services/InlineController'
import { LocalStorage } from './services/LocalStorageProvider'
import {
    CODY_ACCESS_TOKEN_SECRET,
    InMemorySecretStorage,
    SecretStorage,
    VSCodeSecretStorage,
} from './services/SecretStorageProvider'

/**
 * Start the extension, watching all relevant configuration and secrets for changes.
 */
export async function start(context: vscode.ExtensionContext): Promise<vscode.Disposable> {
    const secretStorage =
        process.env.CODY_TESTING === 'true' ? new InMemorySecretStorage() : new VSCodeSecretStorage(context.secrets)
    const localStorage = new LocalStorage(context.globalState)
    const rgPath = await getRgPath(context.extensionPath)

    const disposables: vscode.Disposable[] = []

    const { disposable, onConfigurationChange } = await register(
        context,
        await getFullConfig(secretStorage),
        secretStorage,
        localStorage,
        rgPath
    )
    disposables.push(disposable)

    // Re-initialize when configuration or secrets change.
    disposables.push(
        secretStorage.onDidChange(async key => {
            if (key === CODY_ACCESS_TOKEN_SECRET) {
                onConfigurationChange(await getFullConfig(secretStorage))
            }
        }),
        vscode.workspace.onDidChangeConfiguration(async event => {
            if (event.affectsConfiguration('cody')) {
                onConfigurationChange(await getFullConfig(secretStorage))
            }
        })
    )

    return vscode.Disposable.from(...disposables)
}

// Registers commands and webview given the config.
const register = async (
    context: vscode.ExtensionContext,
    initialConfig: ConfigurationWithAccessToken,
    secretStorage: SecretStorage,
    localStorage: LocalStorage,
    rgPath: string
): Promise<{
    disposable: vscode.Disposable
    onConfigurationChange: (newConfig: ConfigurationWithAccessToken) => void
}> => {
    const disposables: vscode.Disposable[] = []

    await updateEventLogger(initialConfig, localStorage)

    // Controller for inline assist
    const commentController = new InlineController(context.extensionPath)
    disposables.push(commentController.get())

    const fixup = new FixupController()
    const controllers = { inline: commentController, task: fixup, fixup }

    const editor = new VSCodeEditor(controllers)
    const workspaceConfig = vscode.workspace.getConfiguration()
    const config = getConfiguration(workspaceConfig)

    const {
        intentDetector,
        codebaseContext,
        chatClient,
        completionsClient,
        guardrails,
        onConfigurationChange: externalServicesOnDidConfigurationChange,
    } = await configureExternalServices(initialConfig, rgPath, editor)

    // Create chat webview
    const chatProvider = new ChatViewProvider(
        context.extensionPath,
        initialConfig,
        chatClient,
        intentDetector,
        codebaseContext,
        guardrails,
        editor,
        secretStorage,
        localStorage,
        rgPath
    )
    disposables.push(chatProvider)

    disposables.push(
        vscode.window.registerWebviewViewProvider('cody.chat', chatProvider, {
            webviewOptions: { retainContextWhenHidden: true },
        }),
        { dispose: () => vscode.commands.executeCommand('setContext', 'cody.activated', false) }
    )

    const executeRecipe = async (recipe: RecipeID, showTab = true): Promise<void> => {
        if (showTab) {
            await vscode.commands.executeCommand('cody.chat.focus')
        }
        await chatProvider.executeRecipe(recipe, '', showTab)
    }

    const webviewErrorMessager = async (error: string): Promise<void> => {
        if (error.includes('rate limit')) {
            const currentTime: number = Date.now()
            const userPref = localStorage.get('rateLimitError')
            // 21600000 is 6h in ms. ex 6 * 60 * 60 * 1000
            if (!userPref || userPref !== 'never' || currentTime - 21600000 >= parseInt(userPref, 10)) {
                const input = await vscode.window.showErrorMessage(error, 'Do not show again', 'Close')
                switch (input) {
                    case 'Do not show again':
                        await localStorage.set('rateLimitError', 'never')
                        break
                    default:
                        // Save current time as a reminder stamp in 6 hours
                        await localStorage.set('rateLimitError', currentTime.toString())
                }
            }
        }
        chatProvider.sendErrorToWebview(error)
    }

    disposables.push(
        vscode.commands.registerCommand('cody.inline.insert', async (copiedText: string) => {
            // Insert copiedText to the current cursor position
            await vscode.commands.executeCommand('editor.action.insertSnippet', {
                snippet: copiedText,
            })
        }),
        // Inline Assist Provider
        vscode.commands.registerCommand('cody.comment.add', async (comment: vscode.CommentReply) => {
            const isFixMode = /^\/f(ix)?\s/i.test(comment.text.trimStart())
            await commentController.chat(comment, isFixMode)
            await chatProvider.executeRecipe(isFixMode ? 'fixup' : 'inline-chat', comment.text, false)
            logEvent(`CodyVSCodeExtension:inline-assist:${isFixMode ? 'fixup' : 'chat'}`)
        }),
        vscode.commands.registerCommand('cody.comment.delete', (thread: vscode.CommentThread) => {
            commentController.delete(thread)
        }),
        // Access token - this is only used in configuration tests
        vscode.commands.registerCommand('cody.set-access-token', async (args: any[]) => {
            if (args?.length && (args[0] as string)) {
                await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, args[0])
            }
        }),
        vscode.commands.registerCommand('cody.delete-access-token', async () => {
            await chatProvider.logout()
        }),
        // Commands
        vscode.commands.registerCommand('cody.focus', () => vscode.commands.executeCommand('cody.chat.focus')),
        vscode.commands.registerCommand('cody.settings', () => chatProvider.setWebviewView('settings')),
        vscode.commands.registerCommand('cody.history', () => chatProvider.setWebviewView('history')),
        vscode.commands.registerCommand('cody.interactive.clear', async () => {
            await chatProvider.clearAndRestartSession()
            chatProvider.setWebviewView('chat')
        }),
        vscode.commands.registerCommand('cody.recipe.explain-code', () => executeRecipe('explain-code-detailed')),
        vscode.commands.registerCommand('cody.recipe.explain-code-high-level', () =>
            executeRecipe('explain-code-high-level')
        ),
        vscode.commands.registerCommand('cody.recipe.generate-unit-test', () => executeRecipe('generate-unit-test')),
        vscode.commands.registerCommand('cody.recipe.generate-docstring', () => executeRecipe('generate-docstring')),
        vscode.commands.registerCommand('cody.recipe.fixup', () => executeRecipe('fixup')),
        vscode.commands.registerCommand('cody.recipe.translate-to-language', () =>
            executeRecipe('translate-to-language')
        ),
        vscode.commands.registerCommand('cody.recipe.git-history', () => executeRecipe('git-history')),
        vscode.commands.registerCommand('cody.recipe.improve-variable-names', () =>
            executeRecipe('improve-variable-names')
        ),
        vscode.commands.registerCommand('cody.recipe.find-code-smells', () => executeRecipe('find-code-smells')),
        vscode.commands.registerCommand('cody.recipe.context-search', () => executeRecipe('context-search')),
        vscode.commands.registerCommand('cody.recipe.optimize-code', () => executeRecipe('optimize-code')),
        // Register URI Handler for resolving token sending back from sourcegraph.com
        vscode.window.registerUriHandler({
            handleUri: async (uri: vscode.Uri) => {
                const params = new URLSearchParams(uri.query)
                let serverEndpoint = DOTCOM_URL.href
                if (params.get('type') === 'app') {
                    serverEndpoint = LOCAL_APP_URL.href
                }
                await workspaceConfig.update('cody.serverEndpoint', serverEndpoint, vscode.ConfigurationTarget.Global)
                const token = params.get('code')
                if (token && token.length > 8) {
                    await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, token)
                    const authStatus = await getAuthStatus({
                        serverEndpoint,
                        accessToken: token,
                        customHeaders: config.customHeaders,
                    })
                    await chatProvider.sendLogin(authStatus)
                    if (isLoggedIn(authStatus)) {
                        void vscode.window.showInformationMessage('Token has been retrieved and updated successfully')
                    }
                }
            },
        })
    )

    if (initialConfig.experimentalSuggest) {
        // TODO(sqs): make this listen to config and not just use initialConfig
        const docprovider = new CompletionsDocumentProvider()
        disposables.push(vscode.workspace.registerTextDocumentContentProvider('cody', docprovider))

        const history = new History()
        const completionsProvider = new CodyCompletionItemProvider(
            webviewErrorMessager,
            completionsClient,
            docprovider,
            history
        )
        disposables.push(
            vscode.commands.registerCommand('cody.manual-completions', async () => {
                await completionsProvider.fetchAndShowManualCompletions()
            }),
            vscode.commands.registerCommand('cody.completions.inline.accepted', ({ codyLogId }) => {
                CompletionsLogger.accept(codyLogId)
            }),
            vscode.languages.registerInlineCompletionItemProvider({ scheme: 'file' }, completionsProvider)
        )
    }

    // Initiate inline assist when feature flag is on
    if (initialConfig.experimentalInline) {
        commentController.get().commentingRangeProvider = {
            provideCommentingRanges: (document: vscode.TextDocument) => {
                const lineCount = document.lineCount
                return [new vscode.Range(0, 0, lineCount - 1, 0)]
            },
        }
        disposables.push(
            vscode.commands.registerCommand('cody.recipe.file-flow', () => executeRecipe('file-flow', false))
        )
    }

    if (initialConfig.experimentalGuardrails) {
        const guardrailsProvider = new GuardrailsProvider(guardrails, editor)
        disposables.push(
            vscode.commands.registerCommand('cody.guardrails.debug', async () => {
                await guardrailsProvider.debugEditorSelection()
            })
        )
    }
    // Register task view and non-stop cody command when feature flag is on
    if (initialConfig.experimentalNonStop || process.env.CODY_TESTING === 'true') {
        disposables.push(vscode.window.registerTreeDataProvider('cody.fixup.tree.view', fixup.getTaskView()))
        disposables.push(
            vscode.commands.registerCommand('cody.recipe.non-stop', async () => {
                await chatProvider.executeRecipe('non-stop', '', false)
            })
        )
        await vscode.commands.executeCommand('setContext', 'cody.nonstop.fixups.enabled', true)
    }

    return {
        disposable: vscode.Disposable.from(...disposables),
        onConfigurationChange: newConfig => {
            chatProvider.onConfigurationChange(newConfig)
            externalServicesOnDidConfigurationChange(newConfig)
        },
    }
}

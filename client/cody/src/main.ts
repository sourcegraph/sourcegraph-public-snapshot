import * as vscode from 'vscode'

import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Configuration, ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { ChatViewProvider } from './chat/ChatViewProvider'
import { CODY_FEEDBACK_URL } from './chat/protocol'
import { CodyCompletionItemProvider } from './completions'
import { CompletionsDocumentProvider } from './completions/docprovider'
import { History } from './completions/history'
import * as CompletionsLogger from './completions/logger'
import { ManualCompletionService } from './completions/manual'
import { createProviderConfig as createAnthropicProviderConfig } from './completions/providers/anthropic'
import { ProviderConfig } from './completions/providers/provider'
import { createProviderConfig as createUnstableCodeGenProviderConfig } from './completions/providers/unstable-codegen'
import { createProviderConfig as createUnstableHuggingFaceProviderConfig } from './completions/providers/unstable-huggingface'
import { getConfiguration, getFullConfig, migrateConfiguration } from './configuration'
import { VSCodeEditor } from './editor/vscode-editor'
import { logEvent, updateEventLogger, eventLogger } from './event-logger'
import { configureExternalServices } from './external-services'
import { FixupController } from './non-stop/FixupController'
import { showSetupNotification } from './notifications/setup-notification'
import { getRgPath } from './rg'
import { AuthProvider } from './services/AuthProvider'
import { showFeedbackSupportQuickPick } from './services/FeedbackOptions'
import { GuardrailsProvider } from './services/GuardrailsProvider'
import { InlineController } from './services/InlineController'
import { LocalStorage } from './services/LocalStorageProvider'
import {
    CODY_ACCESS_TOKEN_SECRET,
    InMemorySecretStorage,
    SecretStorage,
    VSCodeSecretStorage,
} from './services/SecretStorageProvider'
import { CodyStatusBar, createStatusBar } from './services/StatusBar'
import { TestSupport } from './test-support'

/**
 * Start the extension, watching all relevant configuration and secrets for changes.
 */
export async function start(context: vscode.ExtensionContext): Promise<vscode.Disposable> {
    await migrateConfiguration()

    const secretStorage =
        process.env.CODY_TESTING === 'true' ? new InMemorySecretStorage() : new VSCodeSecretStorage(context.secrets)
    const localStorage = new LocalStorage(context.globalState)
    const rgPath = await getRgPath(context.extensionPath)

    const disposables: vscode.Disposable[] = []

    const { disposable, onConfigurationChange } = await register(
        context,
        await getFullConfig(secretStorage, localStorage),
        secretStorage,
        localStorage,
        rgPath
    )
    disposables.push(disposable)

    // Re-initialize when configuration or secrets change.
    disposables.push(
        secretStorage.onDidChange(async key => {
            if (key === CODY_ACCESS_TOKEN_SECRET) {
                onConfigurationChange(await getFullConfig(secretStorage, localStorage))
            }
        }),
        vscode.workspace.onDidChangeConfiguration(async event => {
            if (event.affectsConfiguration('cody')) {
                onConfigurationChange(await getFullConfig(secretStorage, localStorage))
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
    // Controller for inline Chat
    const commentController = new InlineController(context.extensionPath)

    const fixup = new FixupController()
    disposables.push(fixup)
    if (TestSupport.instance) {
        TestSupport.instance.fixupController.set(fixup)
    }
    const controllers = { inline: commentController, fixups: fixup }

    const editor = new VSCodeEditor(controllers)
    // Could we use the `initialConfig` instead?
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

    const authProvider = new AuthProvider(initialConfig, secretStorage, localStorage)
    await authProvider.init()

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
        rgPath,
        authProvider
    )
    disposables.push(chatProvider)
    fixup.recipeRunner = chatProvider

    disposables.push(
        vscode.window.registerWebviewViewProvider('cody.chat', chatProvider, {
            webviewOptions: { retainContextWhenHidden: true },
        })
    )

    const executeRecipe = async (recipe: RecipeID, openChatView = true): Promise<void> => {
        await chatProvider.executeRecipe(recipe, '', openChatView)
    }

    const webviewErrorMessenger = async (error: string): Promise<void> => {
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

    const statusBar = createStatusBar()

    disposables.push(
        // Inline Chat Provider
        vscode.commands.registerCommand('cody.comment.add', async (comment: vscode.CommentReply) => {
            const isFixMode = /^\/f(ix)?\s/i.test(comment.text.trimStart())
            await commentController.chat(comment, isFixMode)
            await chatProvider.executeRecipe('inline-chat', comment.text.trimStart(), false)
            logEvent(`CodyVSCodeExtension:inline-assist:${isFixMode ? 'fixup' : 'chat'}`)
        }),
        vscode.commands.registerCommand('cody.comment.delete', (thread: vscode.CommentThread) => {
            commentController.delete(thread)
        }),
        vscode.commands.registerCommand('cody.inline.new', () =>
            vscode.commands.executeCommand('workbench.action.addComment')
        ),
        // Tests
        // Access token - this is only used in configuration tests
        vscode.commands.registerCommand('cody.test.token', async (args: any[]) => {
            if (args?.length && (args[0] as string)) {
                await secretStorage.store(CODY_ACCESS_TOKEN_SECRET, args[0])
            }
        }),
        // Auth
        vscode.commands.registerCommand('cody.auth.signin', () => authProvider.signinMenu()),
        vscode.commands.registerCommand('cody.auth.signout', () => authProvider.signoutMenu()),
        vscode.commands.registerCommand('cody.auth.support', () => showFeedbackSupportQuickPick()),
        // Commands
        vscode.commands.registerCommand('cody.interactive.clear', async () => {
            await chatProvider.clearAndRestartSession()
            chatProvider.setWebviewView('chat')
        }),
        vscode.commands.registerCommand('cody.inline.insert', async (copiedText: string) => {
            // Insert copiedText to the current cursor position
            await vscode.commands.executeCommand('editor.action.insertSnippet', {
                snippet: copiedText,
            })
        }),
        vscode.commands.registerCommand('cody.focus', () => vscode.commands.executeCommand('cody.chat.focus')),
        vscode.commands.registerCommand('cody.settings.user', () => chatProvider.setWebviewView('settings')),
        vscode.commands.registerCommand('cody.settings.extension', () =>
            vscode.commands.executeCommand('workbench.action.openSettings', { query: '@ext:sourcegraph.cody-ai' })
        ),
        vscode.commands.registerCommand('cody.history', () => chatProvider.setWebviewView('history')),
        vscode.commands.registerCommand('cody.history.clear', async () => {
            await chatProvider.clearHistory()
        }),
        // Recipes
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
        vscode.commands.registerCommand('cody.recipe.inline-touch', () => executeRecipe('inline-touch', false)),
        vscode.commands.registerCommand('cody.recipe.find-code-smells', () => executeRecipe('find-code-smells')),
        vscode.commands.registerCommand('cody.recipe.context-search', () => executeRecipe('context-search')),

        // Register URI Handler (vscode://sourcegraph.cody-ai)
        vscode.window.registerUriHandler({
            handleUri: async (uri: vscode.Uri) => {
                await authProvider.tokenCallbackHandler(uri, config.customHeaders)
            },
        }),
        statusBar,
        // Walkthrough / Support
        vscode.commands.registerCommand('cody.feedback', () =>
            vscode.env.openExternal(vscode.Uri.parse(CODY_FEEDBACK_URL.href))
        ),
        vscode.commands.registerCommand('cody.welcome', async () => {
            // Hack: We have to run this twice to force VS Code to register the walkthrough
            // Open issue: https://github.com/microsoft/vscode/issues/186165
            await vscode.commands.executeCommand('workbench.action.openWalkthrough')
            return vscode.commands.executeCommand(
                'workbench.action.openWalkthrough',
                'sourcegraph.cody-ai#welcome',
                false
            )
        }),
        vscode.commands.registerCommand('cody.welcome-mock', () =>
            vscode.commands.executeCommand('workbench.action.openWalkthrough', 'sourcegraph.cody-ai#welcome', false)
        ),
        vscode.commands.registerCommand('cody.walkthrough.showLogin', () =>
            vscode.commands.executeCommand('workbench.view.extension.cody')
        ),
        vscode.commands.registerCommand('cody.walkthrough.showChat', () => chatProvider.setWebviewView('chat')),
        vscode.commands.registerCommand('cody.walkthrough.showFixup', () => chatProvider.setWebviewView('recipes')),
        vscode.commands.registerCommand('cody.walkthrough.showExplain', () => chatProvider.setWebviewView('recipes')),
        vscode.commands.registerCommand('cody.walkthrough.enableInlineChat', async () => {
            await workspaceConfig.update('cody.inlineChat', true, vscode.ConfigurationTarget.Global)
            // Open VSCode setting view. Provides visual confirmation that the setting is enabled.
            return vscode.commands.executeCommand('workbench.action.openSettings', {
                query: 'cody.inlineChat.enabled',
                openToSide: true,
            })
        })
    )

    let completionsProvider: vscode.Disposable | null = null
    if (initialConfig.autocomplete) {
        completionsProvider = createCompletionsProvider(
            config,
            webviewErrorMessenger,
            completionsClient,
            statusBar,
            codebaseContext
        )
    }

    // Create a disposable to clean up completions when the extension reloads.
    const disposeCompletions: vscode.Disposable = {
        dispose: () => {
            completionsProvider?.dispose()
        },
    }
    disposables.push(disposeCompletions)

    vscode.workspace.onDidChangeConfiguration(event => {
        if (event.affectsConfiguration('cody.autocomplete')) {
            const config = getConfiguration(vscode.workspace.getConfiguration())

            if (!config.autocomplete) {
                completionsProvider?.dispose()
                completionsProvider = null
                return
            }

            if (completionsProvider !== null) {
                // If completions are already initialized and still enabled, we
                // need to reset the completion provider.
                completionsProvider.dispose()
            }
            completionsProvider = createCompletionsProvider(
                config,
                webviewErrorMessenger,
                completionsClient,
                statusBar,
                codebaseContext
            )
        }
    })

    // Initiate inline chat when feature flag is on
    if (!initialConfig.inlineChat) {
        commentController.dispose()
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
        fixup.register()
        await vscode.commands.executeCommand('setContext', 'cody.nonstop.fixups.enabled', true)
    }

    await showSetupNotification(initialConfig, localStorage)
    void vscode.commands.executeCommand('setContext', 'cody.test.inProgress', process.env.CODY_TESTING === 'true')
    return {
        disposable: vscode.Disposable.from(...disposables),
        onConfigurationChange: newConfig => {
            chatProvider.onConfigurationChange(newConfig)
            externalServicesOnDidConfigurationChange(newConfig)
            if (eventLogger) {
                eventLogger.onConfigurationChange(vscode.workspace.getConfiguration())
            }
        },
    }
}

function createCompletionsProvider(
    config: Configuration,
    webviewErrorMessenger: (error: string) => Promise<void>,
    completionsClient: SourcegraphNodeCompletionsClient,
    statusBar: CodyStatusBar,
    codebaseContext: CodebaseContext
): vscode.Disposable {
    const disposables: vscode.Disposable[] = []

    const documentProvider = new CompletionsDocumentProvider()
    disposables.push(vscode.workspace.registerTextDocumentContentProvider('cody', documentProvider))

    const history = new History()
    const manualCompletionService = new ManualCompletionService(
        webviewErrorMessenger,
        completionsClient,
        documentProvider,
        history,
        codebaseContext
    )
    const providerConfig = createCompletionProviderConfig(config, webviewErrorMessenger, completionsClient)
    const completionsProvider = new CodyCompletionItemProvider({
        providerConfig,
        history,
        statusBar,
        codebaseContext,
        isCompletionsCacheEnabled: config.autocompleteAdvancedCache,
        isEmbeddingsContextEnabled: config.autocompleteAdvancedEmbeddings,
    })

    disposables.push(
        vscode.commands.registerCommand('cody.manual-completions', async () => {
            await manualCompletionService.fetchAndShowManualCompletions()
        }),
        vscode.commands.registerCommand('cody.autocomplete.inline.accepted', ({ codyLogId, codyLines }) => {
            CompletionsLogger.accept(codyLogId, codyLines)
        }),
        vscode.languages.registerInlineCompletionItemProvider('*', completionsProvider)
    )
    return {
        dispose: () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        },
    }
}

function createCompletionProviderConfig(
    config: Configuration,
    webviewErrorMessenger: (error: string) => Promise<void>,
    completionsClient: SourcegraphNodeCompletionsClient
): ProviderConfig {
    let providerConfig: null | ProviderConfig = null
    switch (config.autocompleteAdvancedProvider) {
        case 'unstable-codegen': {
            if (config.autocompleteAdvancedServerEndpoint !== null) {
                providerConfig = createUnstableCodeGenProviderConfig({
                    serverEndpoint: config.autocompleteAdvancedServerEndpoint,
                })
            }

            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            webviewErrorMessenger(
                'Provider `unstable-codegen` can not be used without configuring `cody.autocomplete.advanced.serverEndpoint`. Falling back to `anthropic`.'
            )
            break
        }
        case 'unstable-huggingface': {
            if (config.autocompleteAdvancedServerEndpoint !== null) {
                providerConfig = createUnstableHuggingFaceProviderConfig({
                    serverEndpoint: config.autocompleteAdvancedServerEndpoint,
                    accessToken: config.autocompleteAdvancedAccessToken,
                })
            }

            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            webviewErrorMessenger(
                'Provider `unstable-huggingface` can not be used without configuring `cody.autocomplete.advanced.serverEndpoint`. Falling back to `anthropic`.'
            )
            break
        }
    }
    if (providerConfig) {
        return providerConfig
    }

    return createAnthropicProviderConfig({
        completionsClient,
        contextWindowTokens: 2048,
    })
}

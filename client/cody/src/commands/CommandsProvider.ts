import * as vscode from 'vscode'

import { ChatViewProvider } from './ChatViewProvider'
import { CompletionsDocumentProvider } from './CompletionsDocumentProvider'
import { CODY_ACCESS_TOKEN_SECRET, ConfigurationUseContext, getConfiguration } from './configuration'

export const CommandsProvider = async (context: vscode.ExtensionContext): Promise<void> => {
    const documentProvider = new CompletionsDocumentProvider()
    const config = getConfiguration(vscode.workspace.getConfiguration())
    const accessToken = await context.secrets.get(CODY_ACCESS_TOKEN_SECRET)

    const useContext: ConfigurationUseContext = config.useContext

    // TODO
    // let rgPath = await getRgPath(context.extensionPath);
    // if (!rgPath) {
    //   rgPath = 'rg';
    //   vscode.window.showWarningMessage(
    //     'Did not find bundled `rg` (if running in development, you probably need to run scripts/download-rg.sh). Falling back to the `rg` on $PATH.'
    //   );
    // }

    const chatProvider = new ChatViewProvider(
        config.codebase,
        context.extensionPath,
        config.serverEndpoint,
        accessToken || '',
        config.embeddingsEndpoint,
        useContext,
        config.debug,
        context.extensionPath, // TODO Replace with rgPath
        context.secrets
    )
    vscode.window.registerWebviewViewProvider('cody.chat', chatProvider)
    const disposables: vscode.Disposable[] = []
    disposables.push(
        vscode.commands.registerCommand('sourcegraph.cody.toggleEnabled', async () => {
            const config = vscode.workspace.getConfiguration()
            await config.update(
                'sourcegraph.cody.enable',
                !config.get('sourcegraph.cody.enable'),
                vscode.ConfigurationTarget.Global
            )
        }),
        vscode.workspace.registerTextDocumentContentProvider('codegen', documentProvider),
        vscode.languages.registerHoverProvider({ scheme: 'codegen' }, documentProvider),
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
    // TODO
    // if (config.experimentalSuggest) {
    //   vscode.commands.registerCommand('cody.experimental.suggest', async () => {
    //     await fetchAndShowCompletions(
    //       wsCompletionsClient,
    //       documentProvider,
    //       history
    //     );
    //   });
    // }

    await vscode.commands.executeCommand('setContext', 'sourcegraph.cody.activated', true)

    const executeRecipe = async (recipe: string): Promise<void> => {
        await vscode.commands.executeCommand('cody.chat.focus')
        await chatProvider.executeRecipe(recipe)
    }

    vscode.Disposable.from(...disposables)
}

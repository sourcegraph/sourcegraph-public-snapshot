import * as vscode from 'vscode'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/chat/recipes/prompt-mixin'

import { CommandsProvider } from './command/CommandsProvider'
import { ExtensionApi } from './extension-api'

export function activate(context: vscode.ExtensionContext): Promise<ExtensionApi> {
    console.log('Cody extension activated')

    PromptMixin.add(languagePromptMixin(vscode.env.language))

    // Register commands and webview
    return CommandsProvider(context)
}

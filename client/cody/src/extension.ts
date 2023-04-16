import * as vscode from 'vscode'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/prompt/prompt-mixin'

import { CommandsProvider } from './command/CommandsProvider'
import { ExtensionApi } from './extension-api'

export function activate(context: vscode.ExtensionContext): Promise<ExtensionApi> {
    console.log('Cody extension activated')

    PromptMixin.add(languagePromptMixin(vscode.env.language))

    if (process.env.CODY_FOCUS_ON_STARTUP) {
        setTimeout(() => {
            void vscode.commands.executeCommand('cody.chat.focus')
        }, 250)
    }

    // Register commands and webview
    return CommandsProvider(context)
}

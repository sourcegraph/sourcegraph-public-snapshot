import * as vscode from 'vscode'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/prompt/prompt-mixin'

import { ExtensionApi } from './extension-api'
import { start } from './main'

export function activate(context: vscode.ExtensionContext): ExtensionApi {
    PromptMixin.add(languagePromptMixin(vscode.env.language))

    if (process.env.CODY_FOCUS_ON_STARTUP) {
        setTimeout(() => {
            void vscode.commands.executeCommand('cody.chat.focus')
        }, 250)
    }

    start(context)
        .then(disposable => context.subscriptions.push(disposable))
        .catch(error => console.error(error))

    return new ExtensionApi()
}

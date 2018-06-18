'use strict'
import * as vscode from 'vscode'

import * as lspMain from './lsp/src/main'

// activate is called when the extension is activated.
export function activate(context: vscode.ExtensionContext) {
    lspMain.activate(context)
}

export function deactivate() {
    // no-op
}

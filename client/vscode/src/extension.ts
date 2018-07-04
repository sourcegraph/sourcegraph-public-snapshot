'use strict'
import * as vscode from 'vscode'

import * as lspMain from './main'

// activate is called when the extension is activated.
export async function activate(context: vscode.ExtensionContext): Promise<void> {
    const logger = vscode.window.createOutputChannel('Sourcegraph')
    logger.appendLine('Activating')
    try {
        await lspMain.activate({ context, logger })
    } catch (e) {
        logger.appendLine('Activation failed: ' + e)
    }
}

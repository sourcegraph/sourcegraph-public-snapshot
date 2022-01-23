import * as vscode from 'vscode'

import { invalidateClient } from '../backend/requestGraphQl'
import { VSCEStateMachine } from '../state'

/**
 * Listens for Sourcegraph URL or access token changes and invalidates the GraphQL client
 * to prevent data "contamination" (e.g. sending private repo names to Cloud instance).
 */
export function invalidateContextOnSettingsChange({
    context,
    stateMachine,
}: {
    context: vscode.ExtensionContext
    stateMachine: VSCEStateMachine
}): void {
    function disposeAllResources(): void {
        for (const subscription of context.subscriptions) {
            subscription.dispose()
        }
    }

    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(config => {
            if (config.affectsConfiguration('sourcegraph.accessToken')) {
                invalidateClient()
                disposeAllResources()
                stateMachine.emit({ type: 'access_token_change' })
                vscode.window
                    .showInformationMessage(
                        'Restart VS Code to use the Sourcegraph extension after access token change.'
                    )
                    .then(
                        () => {},
                        () => {}
                    )
            }
            if (config.affectsConfiguration('sourcegraph.url')) {
                invalidateClient()
                disposeAllResources()
                stateMachine.emit({ type: 'sourcegraph_url_change' })
                vscode.window
                    .showInformationMessage('Restart VS Code to use the Sourcegraph extension after URL change.')
                    .then(
                        () => {},
                        () => {}
                    )
            }
        })
    )
}

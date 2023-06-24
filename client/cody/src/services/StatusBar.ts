import * as vscode from 'vscode'

import type { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { getConfiguration } from '../configuration'

import { FeedbackOptionItems } from './FeedbackOptions'

export interface CodyStatusBar {
    dispose(): void
    startLoading(label: string): () => void
}

const DEFAULT_TEXT = '$(cody-logo-heavy)'
const DEFAULT_TOOLTIP = 'Cody Settings'

export function createStatusBar(): CodyStatusBar {
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right)
    statusBarItem.text = DEFAULT_TEXT
    statusBarItem.tooltip = DEFAULT_TOOLTIP
    statusBarItem.command = 'cody.status-bar.interacted'
    statusBarItem.show()

    const command = vscode.commands.registerCommand(statusBarItem.command, async () => {
        const workspaceConfig = vscode.workspace.getConfiguration()
        const config = getConfiguration(workspaceConfig)

        function createFeatureToggle(
            name: string,
            description: string,
            detail: string,
            setting: string,
            getValue: (config: Configuration) => boolean,
            requiresReload: boolean = false
        ): vscode.QuickPickItem & { onSelect: () => Promise<void> } {
            const isEnabled = getValue(config)
            return {
                label: (isEnabled ? '$(check) ' : '') + name,
                description,
                detail,
                onSelect: async () => {
                    await workspaceConfig.update(setting, !isEnabled, vscode.ConfigurationTarget.Global)

                    const info = name + ' ' + (isEnabled ? 'disabled' : 'enabled') + '.'
                    const response = await (requiresReload
                        ? vscode.window.showInformationMessage(info, 'Reload Window')
                        : vscode.window.showInformationMessage(info))

                    if (response === 'Reload Window') {
                        await vscode.commands.executeCommand('workbench.action.reloadWindow')
                    }
                },
            }
        }

        const option = await vscode.window.showQuickPick(
            // These description should stay in sync with the settings in package.json
            [
                { label: 'enable/disable features', kind: vscode.QuickPickItemKind.Separator },
                createFeatureToggle(
                    'Code Autocomplete',
                    'Beta',
                    'Enables inline code suggestions in your editor',
                    'cody.autocomplete.enabled',
                    c => c.autocomplete
                ),
                createFeatureToggle(
                    'Inline Assist',
                    'Beta',
                    'An inline way to explicitly ask questions and propose modifications to code',
                    'cody.experimental.inline',
                    c => c.experimentalInline
                ),

                createFeatureToggle(
                    'Chat Suggestions',
                    'Experimental',
                    'Adds suggestions of possible relevant messages in the chat window',
                    'cody.experimental.chatPredictions',
                    c => c.experimentalChatPredictions,
                    true
                ),
                { label: 'settings', kind: vscode.QuickPickItemKind.Separator },
                {
                    label: '$(gear) Settings',
                    detail: 'Configure your Cody settings.',
                    async onSelect(): Promise<void> {
                        await vscode.commands.executeCommand('cody.settings.extension')
                    },
                },
                { label: 'feedback', kind: vscode.QuickPickItemKind.Separator },
                ...FeedbackOptionItems,
            ],
            {
                placeHolder: 'Select an option',
                matchOnDescription: true,
            }
        )

        if (option && 'onSelect' in option) {
            option.onSelect().catch(console.error)
        }
    })

    // Reference counting to ensure loading states are handled consistently across different
    // features
    // TODO: Ensure the label is always set to the right value too.
    let openLoadingLeases = 0

    return {
        startLoading(label: string) {
            openLoadingLeases++
            statusBarItem.text = '$(loading~spin)'
            statusBarItem.tooltip = label

            let didClose = false
            return () => {
                if (didClose) {
                    return
                }
                didClose = true

                openLoadingLeases--
                if (openLoadingLeases === 0) {
                    statusBarItem.text = DEFAULT_TEXT
                    statusBarItem.tooltip = DEFAULT_TOOLTIP
                }
            }
        },
        dispose() {
            statusBarItem.dispose()
            command.dispose()
        },
    }
}

import * as vscode from 'vscode'

import type { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { getConfiguration } from './configuration'

export interface CodyStatusBar {
    dispose(): void
    startLoading(label: string): () => void
}

export function createStatusBar(): CodyStatusBar {
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right)
    statusBarItem.text = '$(cody-logo)'
    statusBarItem.command = 'cody.status-bar.interacted'
    statusBarItem.show()

    const command = vscode.commands.registerCommand(statusBarItem.command, async () => {
        const workspaceConfig = vscode.workspace.getConfiguration()
        const config = getConfiguration(workspaceConfig)

        function createFeatureToggle(
            name: string,
            setting: string,
            getValue: (config: Configuration) => boolean
        ): vscode.QuickPickItem & { onSelect: () => Promise<void> } {
            const isEnabled = getValue(config)
            return {
                label: (isEnabled ? 'Disable' : 'Enable') + ' ' + name,
                onSelect: async () => {
                    await workspaceConfig.update(setting, !isEnabled, vscode.ConfigurationTarget.Global)
                    await vscode.window.showInformationMessage(name + ' ' + (isEnabled ? 'disabled' : 'enabled') + '.')
                },
            }
        }

        const option = await vscode.window.showQuickPick(
            [
                createFeatureToggle('Code Completions (Beta)', 'cody.completions', c => c.completions),
                createFeatureToggle('Code Inline Assist (Beta)', 'cody.experimental.inline', c => c.experimentalInline),
               
                createFeatureToggle(
                    'Chat Suggestions (Experimental)',
                    'cody.experimental.chatPredictions',
                    c => c.experimentalChatPredictions
                ),
                { label: '', kind: vscode.QuickPickItemKind.Separator },
                {
                    label: '$(feedback) Share Feedback',
                    async onSelect(): Promise<void> {
                        await vscode.env.openExternal(
                            vscode.Uri.parse(
                                'https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/vscode'
                            )
                        )
                    },
                },
            ],
            {
                placeHolder: 'Select an option',
            }
        )

        if (option && 'onSelect' in option) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            option.onSelect()
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
                    statusBarItem.text = '$(cody-logo)'
                    statusBarItem.tooltip = undefined
                }
            }
        },
        dispose() {
            statusBarItem.dispose()
            command.dispose()
        },
    }
}

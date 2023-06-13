import * as vscode from 'vscode'

import type { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { getConfiguration } from '../configuration'

export interface CodyStatusBar {
    dispose(): void
    startLoading(label: string): () => void
}

const DEFAULT_TEXT = '$(cody-logo)'
const DEFAULT_TOOLTIP = 'Cody Features Toggle'

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
            getValue: (config: Configuration) => boolean
        ): vscode.QuickPickItem & { onSelect: () => Promise<void> } {
            const isEnabled = getValue(config)
            return {
                label: (isEnabled ? '$(check) ' : '') + name,
                description,
                detail,
                onSelect: async () => {
                    await workspaceConfig.update(setting, !isEnabled, vscode.ConfigurationTarget.Global)
                    await vscode.window.showInformationMessage(name + ' ' + (isEnabled ? 'disabled' : 'enabled') + '.')
                },
            }
        }

        const option = await vscode.window.showQuickPick(
            // These description should stay in sync with the settings in package.json
            [
                createFeatureToggle(
                    'Code Completions',
                    'Beta',
                    'Experimental Cody completions in your editor',
                    'cody.completions',
                    c => c.completions
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
                    c => c.experimentalChatPredictions
                ),
                { label: 'cody feedback', kind: vscode.QuickPickItemKind.Separator },
                {
                    label: '$(feedback) Share Feedback',
                    detail: 'Ideas or frustrations — we’d love to hear your feedback',
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
                matchOnDescription: true,
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

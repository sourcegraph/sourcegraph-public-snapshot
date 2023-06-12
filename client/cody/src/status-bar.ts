import * as vscode from 'vscode'

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
        const res = await vscode.window.showQuickPick(
            [
                { label: '$(check) Disable Code Completions (Beta)' },
                { label: '$(circle-large-outline) Enable Code Fixups (Beta)' },
                { label: '$(check) Enable Chat Suggestions (Experimental)' },
                { label: '', kind: vscode.QuickPickItemKind.Separator },
                { label: '$(cody-logo) Open Documentation' },
                { label: '$(feedback) Send Feedback' },
            ],
            {
                placeHolder: 'Select an option',
            }
        )

        // const res = await vscode.window.showInformationMessage('Do you want to disable Cody Completions?', 'Disable')
        console.log({ res })
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

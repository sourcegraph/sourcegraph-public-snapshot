import * as vscode from 'vscode'

import { CODY_DOC_URL, CODY_FEEDBACK_URL, DISCORD_URL } from '../chat/protocol'

export const FeedbackOptionItems = [
    {
        label: '$(feedback) Cody Feedback',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse(CODY_FEEDBACK_URL.href))
        },
    },
    {
        label: '$(remote-explorer-documentation) Cody Documentation',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse(CODY_DOC_URL.href))
        },
    },
    {
        label: '$(organization) Cody Discord Channel',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse(DISCORD_URL.href))
        },
    },
]

const FeedbackQuickPickOptions = { title: 'Cody Feedback & Support', placeholder: 'Choose an option' }

export const showFeedbackSupportQuickPick = async (): Promise<void> => {
    const selectedItem = await vscode.window.showQuickPick(FeedbackOptionItems, FeedbackQuickPickOptions)
    await selectedItem?.onSelect()
}

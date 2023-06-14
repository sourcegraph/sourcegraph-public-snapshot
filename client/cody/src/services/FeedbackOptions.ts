import * as vscode from 'vscode'

import { CODY_DOC_URL, CODY_FEEDBACK_URL, DISCORD_URL } from '../chat/protocol'

export const FeedbackOptionItems = [
    {
        label: '$(feedback) Feedback',
        detail: 'Have an idea, found a bug, or need help? Let us know.',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse(CODY_FEEDBACK_URL.href))
        },
    },
    {
        label: '$(remote-explorer-documentation) Documentation',
        detail: 'Search the Cody documentation.',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse(CODY_DOC_URL.href))
        },
    },
    {
        label: '$(organization) Discord Channel',
        detail: 'Join our Discord community’s #cody channel.',
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

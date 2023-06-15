import * as vscode from 'vscode'

export const FeedbackOptions = [
    {
        label: '$(feedback) Feedback',
        detail: 'Have an idea, found a bug, or need help? Let us know.',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(
                vscode.Uri.parse(
                    'https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/vscode'
                )
            )
        },
    },
    {
        label: '$(remote-explorer-documentation) Documentation',
        detail: 'Search the Cody documentation.',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse('https://docs.sourcegraph.com/cody'))
        },
    },
    {
        label: '$(organization) Discord Channel',
        detail: 'Join our Discord community’s #cody channel.',
        async onSelect(): Promise<void> {
            await vscode.env.openExternal(vscode.Uri.parse('https://discord.gg/s2qDtYGnAE'))
        },
    },
]

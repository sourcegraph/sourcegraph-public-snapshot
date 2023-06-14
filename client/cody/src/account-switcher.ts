import * as vscode from 'vscode'

import { getConfiguration } from './configuration'

const KNOWN_AUTHENTICATED_INSTANCES = {
    dotcom: 'Sourcegraph.com',
    app: 'Cody App',
} as const

interface PreviousAuthenticatedInstance {
    label: string
    isCurrent: boolean
}

const getPreviouslyAuthenticatedInstances = (): PreviousAuthenticatedInstance[] => {
    const workspaceConfig = vscode.workspace.getConfiguration()
    const config = getConfiguration(workspaceConfig)

    return [
        {
            label: config.serverEndpoint,
            isCurrent: true,
        },
    ]
}

export const showAccountSwitcher = async (): Promise<void> => {
    const option = await vscode.window.showQuickPick(
        [
            ...getPreviouslyAuthenticatedInstances().map(({ label, isCurrent }) => ({
                label: (isCurrent ? '$(check) ' : '') + label,
                picked: isCurrent,
                onSelect: () => {
                    console.log('hii')
                },
            })),
            {
                label: 'App',
                kind: vscode.QuickPickItemKind.Separator,
                onSelect: () => null,
            },
            {
                label: 'Connect to Cody App',
                alwaysShow: true,
                onSelect: () => {
                    console.log('hii')
                },
            },
            {
                label: 'Login',
                kind: vscode.QuickPickItemKind.Separator,
                onSelect: () => null,
            },
            {
                label: 'Login to a Sourcegraph Enterprise Instance',
                alwaysShow: true,
                onSelect: () => {
                    console.log('hii')
                },
            },
            {
                label: 'Login to Sourcegraph.com',
                alwaysShow: true,
                onSelect: () => {
                    console.log('hii')
                },
            },
            {
                label: 'Login with URL and Access Token',
                alwaysShow: true,
                onSelect: () => {
                    console.log('hii')
                },
            },
        ],
        {
            placeHolder: 'Select an option',
            ignoreFocusOut: true,
        }
    )

    if (option && 'onSelect' in option) {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        option.onSelect()
    }
}

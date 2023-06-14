import * as vscode from 'vscode'

interface Action {
    label: string
    onClick: () => Thenable<void>
}

export interface ActionNotification {
    message: string
    options?: vscode.MessageOptions
    actions: Action[]
}

/**
 * Displays a VS Code information message with actions.
 */
export const showActionNotification = async ({ message, options = {}, actions }: ActionNotification): Promise<void> => {
    const response = await vscode.window.showInformationMessage(
        message,
        options,
        ...actions.map(action => action.label)
    )
    const action = actions.find(action => action.label === response)

    if (!action) {
        return
    }

    return action.onClick()
}

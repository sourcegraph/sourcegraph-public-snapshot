import vscode from 'vscode'

export async function displayWarning(warning: string): Promise<void> {
    await vscode.window.showErrorMessage(warning)
}

// Prompt user to add Sourcegraph to their workspace
export async function addSourcegraphToWrokspace(): Promise<void> {
    const addToWorkspace = await vscode.window
        .showInformationMessage('Would you like to add Sourcegraph to your workspace?', 'Yes', 'No')
        .then(answer => answer)
    if (addToWorkspace === 'Yes') {
        console.log(addToWorkspace)
    }
}

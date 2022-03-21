import vscode from 'vscode'

export async function displayWarning(warning: string | null): Promise<void> {
    if (warning) {
        await vscode.window.showErrorMessage(warning)
    }
}

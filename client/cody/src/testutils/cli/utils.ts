import * as vscode from 'vscode'

export const webviewErrorMessager = async (error: string): Promise<void> => {
    console.error(error)

    return Promise.resolve()
}

export function findSubstringPosition(text: string, substring: string): vscode.Position | null {
    const lines = text.split('\n')

    for (let i = 0; i < lines.length; i++) {
        const index = lines[i].indexOf(substring)
        if (index !== -1) {
            return new vscode.Position(i, index)
        }
    }

    return null
}

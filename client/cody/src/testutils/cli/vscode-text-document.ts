import * as vscode from 'vscode'

/**
 * Mocks the VSCode TextDocument class required for the completion provider.
 */
export class TextDocument implements vscode.TextDocument {
    public fileName = ''
    public isUntitled = false
    public languageId = 'typescript'
    public version = 1
    public isDirty = false
    public isClosed = false

    private text: string

    constructor(public uri: vscode.Uri, text: string) {
        this.text = text.replace(/\r\n/gm, '\n') // normalize end of line
    }

    public save(): Thenable<boolean> {
        throw new Error('Method not implemented.')
    }
    public eol = vscode.EndOfLine.LF
    public lineCount = 0

    private get lines(): string[] {
        return this.text.split('\n')
    }

    public lineAt(position: number | vscode.Position): vscode.TextLine {
        const line = typeof position === 'number' ? position : position.line
        const text = this.lines[line]
        return {
            text,
            range: new vscode.Range(line, 0, line, text.length),
        } as vscode.TextLine
    }

    public offsetAt(position: vscode.Position): number {
        const lines = this.text.split('\n')
        let currentOffSet = 0
        for (let i = 0; i < lines.length; i++) {
            const l = lines[i]
            if (position.line === i) {
                if (l.length < position.character) {
                    throw new Error(
                        `Position ${JSON.stringify(position)} is out of range. Line [${i}] only has length ${l.length}.`
                    )
                }
                return currentOffSet + position.character
            }

            currentOffSet += l.length + 1
        }
        throw new Error(
            `Position ${JSON.stringify(position)} is out of range. Document only has ${lines.length} lines.`
        )
    }

    public positionAt(offset: number): vscode.Position {
        const lines = this.text.split('\n')
        let sum = 0
        for (let i = 0; i < lines.length; i++) {
            const str = lines[i]
            sum += str.length + 1
            if (offset <= sum) {
                return new vscode.Position(i, str.length - (sum - offset) + 1)
            }
        }

        throw new Error('Cannot find position!')
    }
    public getText(range?: vscode.Range): string {
        if (!range) {
            return this.text
        }

        const offset = this.offsetAt(range.start)
        const length = this.offsetAt(range.end) - offset
        return this.text.slice(offset, length)
    }
    public getWordRangeAtPosition(position: vscode.Position, regex?: RegExp): vscode.Range {
        throw new Error('Method not implemented.')
    }
    public validateRange(range: vscode.Range): vscode.Range {
        throw new Error('Method not implemented.')
    }
    public validatePosition(position: vscode.Position): vscode.Position {
        throw new Error('Method not implemented.')
    }
}

import * as vscode from 'vscode'

import { Completion } from '.'

// FIXME: When OpenAI's logit_bias uses a more precise type than 'object',
// specify JSON-able objects as { [prop: string]: JSONSerialiable | undefined }
export type JSONSerializable = null | string | number | boolean | object | JSONSerializable[]

interface Meta {
    elapsedMillis: number
    suffix: string
    llmOptions: JSONSerializable
}

export interface CompletionGroup {
    lang: string
    completions: Completion[]
    meta?: Meta
}

export class CompletionsDocumentProvider implements vscode.TextDocumentContentProvider {
    private completionsByUri: { [uri: string]: CompletionGroup[] } = {}

    private fireDocumentChanged(uri: vscode.Uri): void {
        this.onDidChangeEmitter.fire(uri)
    }

    public clearCompletions(uri: vscode.Uri): void {
        delete this.completionsByUri[uri.toString()]
        this.fireDocumentChanged(uri)
    }

    public addCompletions(uri: vscode.Uri, lang: string, completions: Completion[], debug?: Meta): void {
        if (!this.completionsByUri[uri.toString()]) {
            this.completionsByUri[uri.toString()] = []
        }

        this.completionsByUri[uri.toString()].push({
            lang,
            completions,
            meta: debug,
        })
        this.fireDocumentChanged(uri)
    }

    public onDidChangeEmitter = new vscode.EventEmitter<vscode.Uri>()
    public onDidChange = this.onDidChangeEmitter.event

    public provideTextDocumentContent(uri: vscode.Uri): string {
        const completionGroups = this.completionsByUri[uri.toString()]
        if (!completionGroups) {
            return 'Loading...'
        }

        return completionGroups
            .map(({ completions, lang }) =>
                completions
                    .map(({ content, stopReason: finishReason }, index) => {
                        const completionText = `\`\`\`${lang}\n${content}\n\`\`\``
                        const headerComponents = [`${index + 1} / ${completions.length}`]
                        if (finishReason) {
                            headerComponents.push(`finish_reason:${finishReason}`)
                        }
                        return headerize(headerComponents.join(', '), 80) + '\n' + completionText
                    })
                    .filter(t => t)
                    .join('\n\n')
            )
            .join('\n\n')
    }
}

function headerize(label: string, width: number): string {
    const prefix = '# ======= '
    let buffer = width - label.length - prefix.length - 1
    if (buffer < 0) {
        buffer = 0
    }
    return `${prefix}${label} ${'='.repeat(buffer)}`
}

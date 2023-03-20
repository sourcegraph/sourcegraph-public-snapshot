import * as openai from 'openai'
import * as vscode from 'vscode'

// FIXME: When OpenAI's logit_bias uses a more precise type than 'object',
// specify JSON-able objects as { [prop: string]: JSONSerialiable | undefined }
export type JSONSerializable = null | string | number | boolean | object | JSONSerializable[]

interface Meta {
    elapsedMillis: number
    prompt: string
    suffix: string
    llmOptions: JSONSerializable
}

export interface CompletionGroup {
    lang: string
    prefixText: string
    completions: openai.CreateCompletionResponse
    meta?: Meta
}

export class CompletionsDocumentProvider implements vscode.TextDocumentContentProvider {
    private completionsByUri: { [uri: string]: CompletionGroup[] } = {}

    private isDebug(): boolean {
        return vscode.workspace.getConfiguration().get<boolean>('cody.debug') === true
    }

    private fireDocumentChanged(uri: vscode.Uri): void {
        this.onDidChangeEmitter.fire(uri)
    }

    public clearCompletions(uri: vscode.Uri): void {
        delete this.completionsByUri[uri.toString()]
        this.fireDocumentChanged(uri)
    }

    public addCompletions(
        uri: vscode.Uri,
        lang: string,
        prefixText: string,
        completions: openai.CreateCompletionResponse,
        debug?: Meta
    ): void {
        if (!this.completionsByUri[uri.toString()]) {
            this.completionsByUri[uri.toString()] = []
        }

        this.completionsByUri[uri.toString()].push({
            lang,
            prefixText,
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
            .map(({ completions, lang, prefixText, meta }) =>
                completions.choices
                    .map(({ text, finish_reason, logprobs }, index) => {
                        if (!text) {
                            return undefined
                        }

                        let completionText = `\`\`\`${lang}\n${prefixText}${text}\n\`\`\``
                        if (this.isDebug() && meta) {
                            completionText =
                                `\`\`\`\n${meta.prompt}\n\`\`\`` +
                                '\n' +
                                completionText +
                                '\n' +
                                `\`\`\`\n${meta.suffix}\n\`\`\``
                        }
                        const headerComponents = [`${index + 1} / ${completions.choices.length}`]
                        if (finish_reason) {
                            headerComponents.push(`finish_reason:${finish_reason}`)
                        }
                        if (logprobs?.token_logprobs) {
                            let total = 0
                            for (const logprob of logprobs.token_logprobs) {
                                total += logprob
                            }
                            headerComponents.push(`mean_logprob: ${total / logprobs.token_logprobs.length}`)
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

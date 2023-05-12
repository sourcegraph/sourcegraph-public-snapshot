import vscode from 'vscode'

import { CompletionLogger } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import {
    CodeCompletionParameters,
    CodeCompletionResponse,
    CompletionParameters,
    Event,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { getConfiguration } from './configuration'

let outputChannel: null | vscode.OutputChannel = null

const workspaceConfig = vscode.workspace.getConfiguration()
const config = getConfiguration(workspaceConfig)
if (config.debug) {
    outputChannel = vscode.window.createOutputChannel('Cody AI by Sourcegraph', 'json')
}

export const logger: CompletionLogger = {
    startCompletion(params: CodeCompletionParameters | CompletionParameters) {
        if (!outputChannel) {
            return undefined
        }

        const start = Date.now()
        const type = 'prompt' in params ? 'code-completion' : 'completion'
        let hasFinished = false
        let lastCompletion = ''

        function onError(error: string): void {
            if (hasFinished) {
                return
            }
            hasFinished = true

            outputChannel!.appendLine(
                JSON.stringify({
                    type,
                    status: 'error',
                    duration: Date.now() - start,
                    params,
                    error,
                })
            )
            outputChannel!.appendLine('')
        }

        function onComplete(result: string | CodeCompletionResponse): void {
            if (hasFinished) {
                return
            }
            hasFinished = true

            outputChannel!.appendLine(
                JSON.stringify({
                    type,
                    status: 'success',
                    duration: Date.now() - start,
                    params,
                    result,
                })
            )
            outputChannel!.appendLine('')
        }

        function onEvents(events: Event[]): void {
            for (const event of events) {
                switch (event.type) {
                    case 'completion':
                        lastCompletion = event.completion
                        break
                    case 'error':
                        onError(event.error)
                        break
                    case 'done':
                        onComplete(lastCompletion)
                        break
                }
            }
        }

        return {
            onError,
            onComplete,
            onEvents,
        }
    },
}

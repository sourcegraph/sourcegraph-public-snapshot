import vscode from 'vscode'

import { CompletionLogger } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import {
    CompletionParameters,
    CompletionResponse,
    Event,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { displayName } from '../package.json'

import { getConfiguration } from './configuration'

const workspaceConfig = vscode.workspace.getConfiguration()
const config = getConfiguration(workspaceConfig)

// Output Channels for General and Debug purposes
export const codyOutputChannel = vscode.window.createOutputChannel(displayName)
export const outputChannel = config.debug
    ? vscode.window.createOutputChannel(displayName + ': Debug', 'json')
    : codyOutputChannel

export const logger: CompletionLogger = {
    startCompletion(params: CompletionParameters) {
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

            outputChannel?.appendLine(
                JSON.stringify({
                    type,
                    status: 'error',
                    duration: Date.now() - start,
                    params,
                    error,
                })
            )
            outputChannel?.appendLine('')
        }

        function onComplete(result: string | CompletionResponse): void {
            if (hasFinished) {
                return
            }
            hasFinished = true

            outputChannel?.appendLine(
                JSON.stringify({
                    type,
                    status: 'success',
                    duration: Date.now() - start,
                    params,
                    result,
                })
            )
            outputChannel?.appendLine('')
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

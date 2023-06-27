import vscode from 'vscode'

import { CompletionLogger } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/client'
import {
    CompletionParameters,
    CompletionResponse,
    Event,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { getConfiguration } from './configuration'

const outputChannel: vscode.OutputChannel = vscode.window.createOutputChannel('Cody AI by Sourcegraph', 'json')

/**
 * Logs text for debugging purposes to the "Cody AI by Sourcegraph" output channel.
 *
 * There are three config settings that alter the behavior of this function.
 * A window refresh may be needed if these settings are changed for the behavior
 * change to take effect.
 * - cody.debug.enabled: toggles debug logging on or off
 * - cody.debug.filter: sets a regex filter that opts-in messages with labels matching the regex
 * - cody.debug.verbose: prints out the text in the `verbose` field of the last argument
 *
 * Usage:
 * debug('label', 'this is a message')
 * debug('label', 'this is a message', 'some', 'args')
 * debug('label', 'this is a message', 'some', 'args', { verbose: 'verbose info goes here' })
 */
export function debug(filterLabel: string, text: string, ...args: unknown[]): void {
    const workspaceConfig = vscode.workspace.getConfiguration()
    const config = getConfiguration(workspaceConfig)

    if (!outputChannel || !config.debugEnable) {
        return
    }

    if (config.debugFilter && !config.debugFilter.test(filterLabel)) {
        return
    }

    if (args.length === 0) {
        outputChannel.appendLine(`${filterLabel}: ${text}`)
        return
    }

    const lastArg = args[args.length - 1]
    if (lastArg && typeof lastArg === 'object' && 'verbose' in lastArg) {
        if (config.debugVerbose) {
            outputChannel.appendLine(
                `${filterLabel}: ${text} ${args.slice(0, -1).join(' ')} ${JSON.stringify(lastArg.verbose, null, 2)}`
            )
        } else {
            outputChannel.appendLine(`${filterLabel}: ${text} ${args.slice(0, -1).join(' ')}`)
        }
        return
    }

    outputChannel.appendLine(`${filterLabel}: ${text} ${args.join(' ')}`)
}

export const logger: CompletionLogger = {
    startCompletion(params: CompletionParameters | {}) {
        const workspaceConfig = vscode.workspace.getConfiguration()
        const config = getConfiguration(workspaceConfig)

        if (!config.debugEnable) {
            return undefined
        }

        const start = Date.now()
        const type = 'prompt' in params ? 'code-completion' : 'messages' in params ? 'completion' : 'code-completion'
        let hasFinished = false
        let lastCompletion = ''

        function onError(error: string): void {
            if (hasFinished) {
                return
            }
            hasFinished = true
            debug(
                'CompletionLogger:onError',
                JSON.stringify({
                    type,
                    status: 'error',
                    duration: Date.now() - start,
                    error,
                }),
                { verbose: { params } }
            )
        }

        function onComplete(result: string | CompletionResponse | string[] | CompletionResponse[]): void {
            if (hasFinished) {
                return
            }
            hasFinished = true

            debug(
                'CompletionLogger:onComplete',
                JSON.stringify({
                    type,
                    status: 'success',
                    duration: Date.now() - start,
                }),
                { verbose: { result, params } }
            )
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

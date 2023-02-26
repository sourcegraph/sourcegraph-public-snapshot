import * as openai from 'openai'
import { WebSocket } from 'ws'

import {
	Completion,
	WSCompletionResponse,
	WSCompletionResponseCompletion,
	WSCompletionsRequest,
} from '@sourcegraph/cody-common'

import { enhanceCompletion, tokenCountToChars, truncateByProbability } from './prompts/common'
import { OpenAIBackend, langKeywordStopStrings, promptPrefixOnly } from './prompts/openai'

const openaiKey = process.env.OPENAI_API_KEY
const openaiConfig = new openai.Configuration({ apiKey: openaiKey })
const cushmanBasic = new OpenAIBackend(
	'cushman:prefix-only',
	openaiConfig,
	{
		model: 'code-cushman-001',
		n: 3,
	},
	{
		numGenerated: 256,
		totalSize: 2048,
	},
	promptPrefixOnly(tokenCountToChars(2048 - 256)),
	langKeywordStopStrings
)

export async function wsHandleGetCompletions(ws: WebSocket, req: WSCompletionsRequest): Promise<void> {
	try {
		const completed = [cushmanBasic.getCompletions(req.args)].map(completionPromise =>
			completionPromise
				.then(({ completions: rawCompletions, debug }) => {
					const completions = rawCompletions.flatMap(rawCompletion => {
						const newCompletions: Completion[] = []
						{
							const logprobLimit = -40
							const { truncatedInsertText } = truncateByProbability(logprobLimit, rawCompletion.logprobs)
							const { insertText, prefixText } = enhanceCompletion(
								req.args.prefix,
								truncatedInsertText,
								[]
							)
							newCompletions.push({
								...rawCompletion,
								label: `${rawCompletion.label}:logprob_limit_${-logprobLimit}`,
								insertText,
								prefixText,
							})
						}
						{
							const { insertText, prefixText } = enhanceCompletion(
								req.args.prefix,
								rawCompletion.insertText,
								[]
							)
							newCompletions.push({
								...rawCompletion,
								label: `${rawCompletion.label}:logprob_nolimit`,
								insertText,
								prefixText,
							})
						}

						return newCompletions
					})

					const response: WSCompletionResponseCompletion = {
						completions,
						requestId: req.requestId,
						kind: 'completion',
						debugInfo: debug,
					}

					return new Promise<void>(resolve => {
						ws.send(JSON.stringify(response), err => {
							if (err) {
								console.error(`failed to send websocket getCompletions response: ${err}`)
								return
							}
							resolve()
						})
					})
				})
				.catch(error => {
					console.error('uncaught error:', error)
				})
		)
		await Promise.allSettled(completed)
		const doneMsg: WSCompletionResponse = {
			requestId: req.requestId,
			kind: 'done',
		}
		ws.send(JSON.stringify(doneMsg))
	} catch (error: any) {
		const errMsg: WSCompletionResponse = { requestId: req.requestId, kind: 'error', error: error.toString() }
		ws.send(JSON.stringify(errMsg))
	}
}

import { Message } from '@sourcegraph/cody-common'

import { AnthropicAPIClient, AnthropicCompletionCallbacks, AnthropicCompletionParams } from './anthropic-api'

const logRequests = ['true', '1'].indexOf(process.env.LOG_CLAUDE?.toLocaleLowerCase() || '') >= 0

export class ClaudeBackend {
	private client: AnthropicAPIClient

	constructor(
		apiKey: string,
		private modelParams: Omit<AnthropicCompletionParams, 'prompt'>,
		private preambleMessages: Message[]
	) {
		this.client = new AnthropicAPIClient(apiKey)
	}

	chat(origMessages: Message[], callbacks: AnthropicCompletionCallbacks): void {
		const messages = [...this.preambleMessages, ...origMessages]

		let lastSpeaker: 'bot' | 'you' | undefined
		for (const msg of messages) {
			if (msg.speaker === lastSpeaker) {
				throw new Error(`duplicate speaker ${lastSpeaker}`)
			}
			lastSpeaker = msg.speaker
		}

		const promptComponents: string[] = []
		for (const msg of messages) {
			promptComponents.push(`\n\n${msg.speaker === 'bot' ? 'Assistant' : 'Human'}: ${msg.text}`)
		}

		if (lastSpeaker === 'you') {
			promptComponents.push('\n\nAssistant:') // Important: no trailing space (affects output quality)
		}

		const prompt = promptComponents.join('')

		if (logRequests) {
			console.log(`REQUEST:\n${prompt}`)
		}

		const anthropicReq = { ...this.modelParams, prompt }
		console.log(`REQ: ${JSON.stringify(anthropicReq, null, '  ')}`)
		this.client.completion(anthropicReq, callbacks)
	}
}

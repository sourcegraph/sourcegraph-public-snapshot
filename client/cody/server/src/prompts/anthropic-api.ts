import fetch from 'node-fetch'

const API_URL = 'https://api.anthropic.com'

export interface AnthropicCompletionParams {
	prompt: string
	temperature: number
	max_tokens_to_sample: number
	stop_sequences: string[]
	top_k: number
	top_p: number
	model: string
}

export interface AnthropicCompletionCallbacks {
	onChange: (text: string) => void
	onComplete: (text: string) => void
	onError: (error: any) => void
}

interface AnthropicCompletionResponse {
	completion: string
	stop_reason: string | null
	stop: string | null
}

const RESPONSE_DATA_PREFIX = 'data: '
const RESPONSE_DONE_PREFIX = `${RESPONSE_DATA_PREFIX}[DONE]`

export class AnthropicAPIClient {
	private headers: {
		Accept: string
		'X-API-Key': string
		'Content-Type': string
		Client: string
	}
	private completionURL: string

	constructor(apiKey: string) {
		this.headers = {
			Accept: 'application/json',
			'X-API-Key': apiKey,
			'Content-Type': 'application/json',
			Client: 'sourcegraph-cody-v1',
		}
		this.completionURL = `${API_URL}/v1/complete`
	}

	async completion(params: AnthropicCompletionParams, callbacks: AnthropicCompletionCallbacks): Promise<void> {
		const response = await fetch(this.completionURL, {
			method: 'POST',
			headers: this.headers,
			body: JSON.stringify({ ...params, stream: true }),
		})

		const onResponseLine = (line: string): void => {
			if (!line.startsWith(RESPONSE_DATA_PREFIX) || line === RESPONSE_DONE_PREFIX) {
				return
			}

			const completionResponse = parseCompletionResponse(line.slice(RESPONSE_DATA_PREFIX.length))
			if (!completionResponse) {
				console.error('Could not parse line:', line)
				return
			}

			if (completionResponse.stop_reason === null) {
				callbacks.onChange(completionResponse.completion)
			} else {
				callbacks.onComplete(completionResponse.completion)
			}
		}

		try {
			let buffer = ''
			for await (const chunk of response.body) {
				buffer += chunk.toString()
				buffer = iterateLines(buffer, onResponseLine)
			}
			if (buffer.length > 0) {
				iterateLines(buffer, onResponseLine)
			}
		} catch (error: any) {
			console.error(error)
			callbacks.onError(error)
		}
	}
}

function parseCompletionResponse(data: string): AnthropicCompletionResponse | null {
	try {
		return JSON.parse(data) as AnthropicCompletionResponse
	} catch {
		return null
	}
}

function iterateLines(buffer: string, onResponseLine: (line: string) => void): string {
	if (buffer.length === 0) {
		return ''
	}

	const lines = buffer.split('\n')
	const lastLine = lines[lines.length - 1]
	const isPartialLastLine = lastLine && lastLine.endsWith(buffer[buffer.length - 1])

	const rest = isPartialLastLine ? lastLine : ''
	const numCompletedLines = isPartialLastLine ? lines.length - 1 : lines.length

	for (const line of lines.slice(0, numCompletedLines)) {
		onResponseLine(line.trim())
	}

	return rest
}

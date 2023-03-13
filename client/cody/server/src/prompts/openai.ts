import { performance } from 'perf_hooks'

import { AxiosResponse } from 'axios'
import { Configuration, CreateCompletionRequest, CreateCompletionResponse, OpenAIApi } from 'openai'

import { Completion, CompletionsArgs, InflatedHistoryItem, LLMDebugInfo, ReferenceInfo } from '@sourcegraph/cody-common'

import { getCharCountLimitedPrefixAtLineBreak, tokenCost, tokenCountToChars } from './common'
import { createCompletion } from './openai-ratelimit'

type RawCompletion = Omit<Completion, 'prefixText'>
export class OpenAIBackend {
	oa: OpenAIApi

	constructor(
		private label: string,
		config: Configuration,
		private defaultCompletionParams: CreateCompletionRequest,
		private contextWindowInfo: {
			totalSize: number
			numGenerated: number
		},
		private createPrompt: (opt: PromptOpt) => string,
		private stopStrings: (uri: string) => string[] | undefined,
		private extractFromResponse_?: (responseText: string) => string
	) {
		this.oa = new OpenAIApi(config)
		this.defaultCompletionParams = {
			max_tokens: this.contextWindowInfo.numGenerated,
			...this.defaultCompletionParams,
		}
	}

	extractFromResponse(text: string): string {
		return this.extractFromResponse_ ? this.extractFromResponse_(text) : text
	}

	async getCompletions({ uri, prefix, history, references }: CompletionsArgs): Promise<{
		debug: LLMDebugInfo
		completions: RawCompletion[]
	}> {
		const startTime = performance.now()
		const prompt = this.createPrompt({
			prefix,
			history,
			references,
			maxTokens: this.contextWindowInfo.totalSize - this.contextWindowInfo.numGenerated,
			tokenCost,
		})
		console.log('# prompt', prompt)
		const opt = {
			...this.defaultCompletionParams,
			stop: this.stopStrings(uri),
		}

		let response: Pick<AxiosResponse<CreateCompletionResponse, any>, 'data'> | undefined
		try {
			// using createCompletion instead of this.oa.createCompletion to ensure openai requests happen serial (see function docs)
			response = await createCompletion(this.oa, {
				...opt,
				prompt,
				logprobs: 1,
			})
		} catch (error) {
			if (error) {
				throw new Error(`OpenAI API error: ${error}`)
			}
		}
		if (!response) {
			throw new Error('OpenAI API repsonser was undefined')
		}

		const endTime = performance.now()
		const debug = {
			prompt,
			llmOptions: opt,
			elapsedMillis: endTime - startTime,
		}
		const completions: RawCompletion[] = []
		for (const choice of response.data.choices) {
			if (!choice.text) {
				continue
			}
			choice.finish_reason
			choice.logprobs
			completions.push({
				label: this.label,
				insertText: this.extractFromResponse(choice.text),
				logprobs: {
					tokens: choice.logprobs?.tokens,
					tokenLogprobs: choice.logprobs?.token_logprobs,
					topLogprobs: choice.logprobs?.top_logprobs,
					textOffset: choice.logprobs?.text_offset,
				},
				finishReason: choice.finish_reason,
			})
		}
		return {
			debug,
			completions,
		}
	}
}

interface PromptOpt {
	prefix: string
	history?: InflatedHistoryItem[]
	references?: ReferenceInfo[]

	maxTokens: number
	tokenCost: (s: string, assumeExtraNewlines?: number) => number
}

/**
 * promptPrefixOnly uses the text preceding the cursor as the prompt, truncating
 * to the nearest complete line that fits in the estimated token window.
 *
 * @param lookback is the lookback window in chars
 * @returns
 */
export const promptPrefixOnly =
	(lookback: number) =>
	({
		prefix,

		maxTokens,
	}: PromptOpt): string => {
		if (lookback > tokenCountToChars(maxTokens)) {
			throw new Error('requested lookback exceeds maxTokens for this model')
		}
		return getCharCountLimitedPrefixAtLineBreak(prefix, lookback)
	}

export function newlineStopString(): string[] {
	return ['\n']
}

export function endCodeBlockStopStrings(): string[] {
	return ['```']
}

export function langKeywordStopStrings(uri: string): string[] | undefined {
	const ext = uri.split('.').pop()?.toLowerCase()
	// switch statement on ext
	switch (ext) {
		case 'ts':
			return ['\nfunction ', '\nclass ', '\nexport ', '\n```']
		case 'go':
			return ['\nfunc ', '\ntype ']
		default:
			console.error(`no stopwords defined for language ${ext}, defaulting to no stopwords`)
			return undefined
	}
}

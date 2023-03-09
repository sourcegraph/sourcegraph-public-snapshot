import { InflatedHistoryItem } from './history'
export * from './history'

// FIXME: When OpenAI's logit_bias uses a more precise type than 'object',
// specify JSON-able objects as { [prop: string]: JSONSerialiable | undefined }
export type JSONSerializable = null | string | number | boolean | object | JSONSerializable[]

export interface ReferenceInfo {
	text: string
	filename: string
}

export interface LLMDebugInfo {
	elapsedMillis: number
	prompt: string
	llmOptions: JSONSerializable
}

export interface Message {
	speaker: 'you' | 'bot'
	text: string
}

export interface ContextMessage extends Message {
	filename?: string
}

/**
 * Each TranscriptChunk corresponds to a sequence of messages that should be considered as a unit during prompt construction.
 * - Typically, `actual` has length 1 and represents the actual message incorporated into the prompt.
 * - `context` is messages that include code snippets fetched as contextual knowledge.
 *    These should not be displayed in the chat GUI.
 * - `display` are messages that should replace `actual` in the chat GUI.
 */
export interface TranscriptChunk {
	actual: Message[]
	context: ContextMessage[]
	display?: Message[]
}

export interface CompletionLogProbs {
	tokens?: string[]
	tokenLogprobs?: number[]
	topLogprobs?: object[]
	textOffset?: number[]
}

export interface Completion {
	/**
	 * The label to display for this completion.
	 */
	label: string
	/**
	 * The text that should be prepended to the insertText to arrive at a "well-formed" (e.g., mostly balanced) completion.
	 */
	prefixText: string

	/**
	 * The text to insert at the point of completion.
	 */
	insertText: string

	/**
	 * Log probabilities of the completion tokens
	 */
	logprobs?: CompletionLogProbs

	/**
	 * The reason the completion terminated
	 */
	finishReason?: string
}
export interface CompletionsArgs {
	uri: string
	prefix: string
	history: InflatedHistoryItem[]
	references: ReferenceInfo[]
}
export interface WSResponse {
	requestId?: number
}
export interface WSCompletionsRequest {
	requestId: number
	kind: 'getCompletions'
	args: CompletionsArgs
}
export interface WSCompletionResponseCompletion extends WSResponse {
	kind: 'completion'
	completions: Completion[]
	debugInfo?: LLMDebugInfo
}
export interface WSCompletionResponseError extends WSResponse {
	kind: 'error'
	error: string
}
export interface WSCompletionResponseMetadata extends WSResponse {
	kind: 'metadata'
	metadata: JSONSerializable
}
export interface WSCompletionResponseDone extends WSResponse {
	kind: 'done'
}
export type WSCompletionResponse =
	| WSCompletionResponseCompletion
	| WSCompletionResponseError
	| WSCompletionResponseMetadata
	| WSCompletionResponseDone

export interface WSChatMessage {
	kind: 'request' | 'response:change' | 'response:complete' | 'response:error'
	requestId: number
}
export interface WSChatRequest extends WSChatMessage {
	kind: 'request'
	messages: Message[]
	metadata?: ChatMetadata
}
export interface ChatMetadata {
	displayMessageLength: number
	recipeId: string | null
}
export type WSChatResponse = WSChatResponseChange | WSChatResponseComplete | WSChatResponseError
export interface WSChatResponseChange extends WSChatMessage {
	kind: 'response:change'
	message: string
}
export interface WSChatResponseComplete extends WSChatMessage {
	kind: 'response:complete'
	message: string
}
export interface WSChatResponseError extends WSChatMessage {
	kind: 'response:error'
	error: string
}

export interface QueryInfo {
	needsCodebaseContext: boolean
	needsCurrentFileContext: boolean
}

export interface Feedback {
	user: string
	sentiment: 'good' | 'bad'
	displayMessages: Message[]
	transcript: TranscriptChunk[]
	feedbackVersion: string
}

export function feedbackToSheetRow({ user, sentiment, displayMessages, transcript, feedbackVersion }: Feedback): {
	[header: string]: string | boolean | number
} {
	return {
		user,
		sentiment,
		displayMessages: JSON.stringify(displayMessages),
		transcript: JSON.stringify(transcript),
		timestamp: new Date().toISOString(),
		feedbackVersion,
	}
}

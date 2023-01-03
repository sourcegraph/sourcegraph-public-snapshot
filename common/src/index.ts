import { InflatedHistoryItem } from './history';
export * from './history';

export interface ReferenceInfo {
	text: string;
	filename: string;
}

export interface LLMDebugInfo {
	elapsedMillis: number;
	prompt: string;
	llmOptions: any;
}

export interface Message {
	speaker: "you" | "bot";
	text: string;
}

export interface CompletionsArgs {
	uri: string;
	prefix: string;
	history: InflatedHistoryItem[];
	references: ReferenceInfo[];
}
export interface WSResponse {
	requestId?: number;
}
export interface WSCompletionsRequest {
	requestId: number
	kind: 'getCompletions'
	args: CompletionsArgs
}
export interface WSCompletionResponseCompletion extends WSResponse {
	kind: 'completion'
	completions: string[]
	debugInfo?: LLMDebugInfo
}
export interface WSCompletionResponseError extends WSResponse {
	kind: 'error'
	error: string
}
export interface WSCompletionResponseMetadata extends WSResponse {
	kind: 'metadata'
	metadata: any
}
export interface WSCompletionResponseDone extends WSResponse {
	kind: 'done'
}
export type WSCompletionResponse = WSCompletionResponseCompletion | WSCompletionResponseError | WSCompletionResponseMetadata | WSCompletionResponseDone;

export interface WSChatMessage {
	kind: 'request' | 'response:change' | 'response:complete' | 'response:error'
	requestId: number;
}
export interface WSChatRequest extends WSChatMessage {
	kind: 'request'
	messages: Message[]
}
export type WSChatResponse = WSChatResponseChange | WSChatResponseComplete | WSChatResponseError;
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
import { Message } from "@sourcegraph/cody-common";

import {
	AnthropicAPIClient,
	AnthropicCompletionCallbacks,
	AnthropicCompletionParams,
} from "./anthropic-api";

export class ClaudeBackend {
	private client: AnthropicAPIClient;

	constructor(
		apiKey: string,
		private modelParams: Omit<AnthropicCompletionParams, "prompt">
	) {
		this.client = new AnthropicAPIClient(apiKey);
	}

	chat(messages: Message[], callbacks: AnthropicCompletionCallbacks): void {
		// basic verification
		let lastSpeaker: "bot" | "you" | undefined;
		for (const msg of messages) {
			if (msg.speaker === lastSpeaker) {
				throw new Error(`duplicate speaker ${lastSpeaker}`);
			}
			lastSpeaker = msg.speaker;
		}
		if (lastSpeaker !== "you") {
			throw new Error("last speaker was not human");
		}

		const promptComponents: string[] = [];
		for (const msg of messages) {
			promptComponents.push(
				`\n\n${msg.speaker === "bot" ? "Assistant" : "Human"}: ${msg.text}`
			);
		}
		promptComponents.push("\n\nAssistant: ");
		const prompt = promptComponents.join("");

		this.client.completion({ ...this.modelParams, prompt }, callbacks);
	}
}

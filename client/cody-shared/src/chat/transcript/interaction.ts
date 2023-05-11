import { ContextMessage } from '../../codebase-context/messages'
import { PromptMixin } from '../../prompt/prompt-mixin'
import { Message } from '../../sourcegraph-api'

import { ChatMessage, InteractionMessage } from './messages'

export interface InteractionJSON {
    humanMessage: InteractionMessage
    assistantMessage: InteractionMessage
    context: ContextMessage[]
    timestamp: string
}

export class Interaction {
    private readonly humanMessage: InteractionMessage
    private assistantMessage: InteractionMessage
    public readonly timestamp: string
    private readonly context: Promise<ContextMessage[]>

    // A sorted list of unique filenames of context files that will be set later.
    private cachedContextFileNames: string[] = []

    constructor(
        humanMessage: InteractionMessage,
        assistantMessage: InteractionMessage,
        context: Promise<ContextMessage[]>,
        timestamp: string = new Date().toISOString()
    ) {
        this.humanMessage = humanMessage
        this.assistantMessage = assistantMessage
        this.timestamp = timestamp

        // This is some hacky behavior: returns a promise that resolves to the same array that was passed,
        // but also caches the context file names in memory as a side effect.
        this.context = context.then(messages => {
            // Extract the context file names from the context messages.
            const contextFileNames = messages
                .map(message => message.fileName)
                .filter((fileName): fileName is string => !!fileName)

            // Cache the context files in memory, so we don't have to block the UI
            // when calling `toChat` by waiting for the context to resolve.
            this.cachedContextFileNames = [...new Set<string>(contextFileNames)].sort((a, b) => a.localeCompare(b))

            return messages
        })
    }

    public getAssistantMessage(): InteractionMessage {
        return this.assistantMessage
    }

    public setAssistantMessage(assistantMessage: InteractionMessage): void {
        this.assistantMessage = assistantMessage
    }

    public async hasContext(): Promise<boolean> {
        const contextMessages = await this.context
        return contextMessages.length > 0
    }

    public async toPrompt(includeContext: boolean): Promise<Message[]> {
        const messages: (ContextMessage | InteractionMessage)[] = [
            PromptMixin.mixInto(this.humanMessage),
            this.assistantMessage,
        ]
        if (includeContext) {
            messages.unshift(...(await this.context))
        }

        return messages.map(message => ({ speaker: message.speaker, text: message.text }))
    }

    /**
     * Converts the interaction to chat message pair: one message from a human, one from an assistant.
     */
    public toChat(): ChatMessage[] {
        return [this.humanMessage, { ...this.assistantMessage, contextFiles: this.cachedContextFileNames }]
    }

    public async toJSON(): Promise<InteractionJSON> {
        return {
            humanMessage: this.humanMessage,
            assistantMessage: this.assistantMessage,
            context: await this.context,
            timestamp: this.timestamp,
        }
    }
}

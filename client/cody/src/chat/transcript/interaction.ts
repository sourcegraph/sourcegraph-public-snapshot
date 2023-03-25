import { ContextMessage } from '../../codebase-context/messages'
import { Message } from '../../sourcegraph-api'

import { ChatMessage, InteractionMessage } from './messages'

export class Interaction {
    private cachedContextFileNames: string[] = []

    constructor(
        private humanMessage: InteractionMessage,
        private assistantMessage: InteractionMessage,
        private context: Promise<ContextMessage[]>
    ) {}

    public getAssistantMessage(): InteractionMessage {
        return this.assistantMessage
    }

    public setAssistantMessage(assistantMessage: InteractionMessage): void {
        this.assistantMessage = assistantMessage
    }

    public async toPrompt(includeContext: boolean): Promise<Message[]> {
        if (includeContext) {
            const contextMessages = await this.context.then(messages => {
                const contextFileNames = messages
                    .map(message => message.fileName)
                    .filter((fileName): fileName is string => !!fileName)

                // Cache the context files so we don't have to block the UI when calling `toChat` by waiting for the context to resolve.
                this.cachedContextFileNames = [...new Set<string>(contextFileNames)].sort((a, b) => a.localeCompare(b))

                return messages
            })
            return [...contextMessages, this.humanMessage, this.assistantMessage].map(toPromptMessage)
        }
        return [this.humanMessage, this.assistantMessage].map(toPromptMessage)
    }

    public toChat(): ChatMessage[] {
        return [this.humanMessage, { ...this.assistantMessage, contextFiles: this.cachedContextFileNames }]
    }
}

function toPromptMessage(interactionOrContextMessage: InteractionMessage | ContextMessage): Message {
    return { speaker: interactionOrContextMessage.speaker, text: interactionOrContextMessage.text }
}

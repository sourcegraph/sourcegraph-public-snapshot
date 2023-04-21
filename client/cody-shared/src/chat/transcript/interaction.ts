import { ContextMessage } from '../../codebase-context/messages'
import { PromptMixin } from '../../prompt/prompt-mixin'
import { Message } from '../../sourcegraph-api'

import { ChatMessage, InteractionMessage } from './messages'

export interface InteractionJSON {
    humanMessage: InteractionMessage
    assistantMessage: InteractionMessage
    context: ContextMessage[]
}

export class Interaction {
    private cachedContextFileNames: string[] = []
    private context: Promise<ContextMessage[]>

    constructor(
        private humanMessage: InteractionMessage,
        private assistantMessage: InteractionMessage,
        context: Promise<ContextMessage[]>
    ) {
        this.context = context.then(messages => {
            const contextFileNames = messages
                .map(message => message.fileName)
                .filter((fileName): fileName is string => !!fileName)

            // Cache the context files so we don't have to block the UI when calling `toChat` by waiting for the context to resolve.
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
        return messages.map(toPromptMessage)
    }

    public toChat(): ChatMessage[] {
        return [this.humanMessage, { ...this.assistantMessage, contextFiles: this.cachedContextFileNames }]
    }

    public async toJSON(): Promise<InteractionJSON> {
        return { humanMessage: this.humanMessage, assistantMessage: this.assistantMessage, context: await this.context }
    }
}

function toPromptMessage(interactionOrContextMessage: InteractionMessage | ContextMessage): Message {
    return { speaker: interactionOrContextMessage.speaker, text: interactionOrContextMessage.text }
}

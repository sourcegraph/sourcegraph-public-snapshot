import { ContextMessage, ContextFile } from '../../codebase-context/messages'
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
    private cachedContextFiles: ContextFile[] = []
    public readonly timestamp: string
    private readonly context: Promise<ContextMessage[]>

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
            const contextFilesMap = messages.reduce((map, { file }) => {
                if (!file?.fileName) {
                    return map
                }
                map[`${file.repoName || 'repo'}@${file?.revision || 'HEAD'}/${file.fileName}`] = file
                return map
            }, {} as { [key: string]: ContextFile })

            // Cache the context files so we don't have to block the UI when calling `toChat` by waiting for the context to resolve.
            this.cachedContextFiles = [
                ...Object.keys(contextFilesMap)
                    .sort((a, b) => a.localeCompare(b))
                    .map((key: string) => contextFilesMap[key]),
            ]

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
        return [this.humanMessage, { ...this.assistantMessage, contextFiles: this.cachedContextFiles }]
    }

    public async toChatPromise(): Promise<ChatMessage[]> {
        await this.context
        return this.toChat()
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

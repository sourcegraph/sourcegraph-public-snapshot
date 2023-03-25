import { CHARS_PER_TOKEN, MAX_AVAILABLE_PROMPT_LENGTH } from '@sourcegraph/cody-shared/src/prompt/constants'
import { getShortTimestamp } from '@sourcegraph/cody-shared/src/timestamp'

import { Message } from '../../sourcegraph-api'

import { Interaction } from './interaction'
import { ChatMessage } from './messages'

export class Transcript {
    private interactions: Interaction[] = []

    public addInteraction(interaction: Interaction | null): void {
        if (!interaction) {
            return
        }
        this.interactions.push(interaction)
    }

    private getLastInteraction(): Interaction | null {
        return this.interactions.length > 0 ? this.interactions[this.interactions.length - 1] : null
    }

    public addAssistantResponse(text: string, displayText: string): void {
        this.getLastInteraction()?.setAssistantMessage({
            speaker: 'assistant',
            text,
            displayText,
            timestamp: getShortTimestamp(),
        })
    }

    public async toPrompt(): Promise<Message[]> {
        const messages: Message[] = []
        for (let index = 0; index < this.interactions.length; index++) {
            const interactionMessages = await this.interactions[index].toPrompt(index === this.interactions.length - 1)
            messages.push(...interactionMessages)
        }
        return truncatePrompt(messages)
    }

    public toChat(): ChatMessage[] {
        return this.interactions.flatMap(interaction => interaction.toChat())
    }

    public reset(): void {
        this.interactions = []
    }
}

function truncatePrompt(messages: Message[]): Message[] {
    const newPromptMessages = []
    let availablePromptTokensBudget = MAX_AVAILABLE_PROMPT_LENGTH
    for (let i = messages.length - 1; i >= 1; i -= 2) {
        const humanMessage = messages[i - 1]
        const botMessage = messages[i]
        const combinedTokensUsage = estimateTokensUsage(humanMessage) + estimateTokensUsage(botMessage)

        // We stop adding pairs of messages once we exceed the available tokens budget.
        if (combinedTokensUsage <= availablePromptTokensBudget) {
            newPromptMessages.push(botMessage, humanMessage)
            availablePromptTokensBudget -= combinedTokensUsage
        } else {
            break
        }
    }

    // Reverse the prompt messages, so they appear in chat order (older -> newer).
    return newPromptMessages.reverse()
}

function estimateTokensUsage(message: Message): number {
    return Math.round(message.text.length / CHARS_PER_TOKEN)
}

import { HNSWLib } from 'langchain/vectorstores/hnswlib'

import { Interaction } from '@sourcegraph/cody-shared/src/chat/transcript/interaction'
import { InteractionMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { ContextSearchOptions } from '@sourcegraph/cody-shared/src/codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '@sourcegraph/cody-shared/src/codebase-context/messages'
// import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'
import { MAX_HUMAN_INPUT_TOKENS } from '@sourcegraph/cody-shared/src/prompt/constants'
import { populateMarkdownContextTemplate } from '@sourcegraph/cody-shared/src/prompt/templates'
import { truncateText } from '@sourcegraph/cody-shared/src/prompt/truncation'

import { CodebaseContexts } from '../constants'

class SlackInteraction {
    public contextMessages: ContextMessage[] = []

    constructor(private humanMessage: InteractionMessage, private assistantMessage: InteractionMessage) {}

    public async updateContextMessagesFromVectorStore(vectorStore: HNSWLib, numResults: number) {
        const docs = await vectorStore.similaritySearch(this.humanMessage.text!, numResults)

        docs.forEach(doc => {
            const contextMessage = getContextMessageWithResponse(
                populateMarkdownContextTemplate(doc.pageContent, doc.metadata.fileName),
                doc.metadata.fileName
            )
            this.contextMessages.push(...contextMessage)
        })
    }

    public async updateContextMessages(
        codebaseContexts: CodebaseContexts,
        codebase: keyof CodebaseContexts,
        contextSearchOptions: ContextSearchOptions
        // intentDetector?: IntentDetector
    ) {
        // const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)
        const isCodebaseContextRequired = true

        if (isCodebaseContextRequired) {
            const contextMessages = await codebaseContexts[codebase].getContextMessages(
                this.humanMessage.text!,
                contextSearchOptions
            )

            this.contextMessages.push(
                ...contextMessages.map(message => {
                    if (message.file) {
                        message.file.fileName = `https://${codebase}/blob/main/${message.file.fileName}`
                    }

                    return message
                })
            )
        }
    }

    public getTranscriptInteraction() {
        return new Interaction(this.humanMessage, this.assistantMessage, Promise.resolve(this.contextMessages))
    }
}

export function getSlackInteraction(humanText: string, assistantText: string = ''): SlackInteraction {
    const text = cleanupMessageForPrompt(humanText)
    const filteredHumanText = truncateText(text, MAX_HUMAN_INPUT_TOKENS)

    return new SlackInteraction(
        { speaker: 'human', text: filteredHumanText },
        { speaker: 'assistant', text: assistantText }
    )
}

export function cleanupMessageForPrompt(text: string, isAssistantMessage = false) {
    // Delete mentions
    const textWithoutMentions = text.replace(/<@[\dA-Z]+>/gm, '').trim()

    // Delete cody-slack filters
    const textWithoutFilters = textWithoutMentions.replace(/channel:([\w-]+)/gm, '').trim()

    if (isAssistantMessage) {
        // Delete "Files used" section
        const filesSectionIndex = textWithoutFilters.lastIndexOf('*Files used*​')

        if (filesSectionIndex !== -1) {
            return textWithoutFilters
                .slice(0, filesSectionIndex)
                .replace(/[|\u00A0\u200B\u200D]/gm, '')
                .replace(/\n+$/gm, '')
        }
    }

    return textWithoutFilters
}

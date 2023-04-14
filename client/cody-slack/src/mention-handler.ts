import { AppMentionEvent } from '@slack/bolt'
import { Message as SlackReplyMessage } from '@slack/web-api/dist/response/ChannelsRepliesResponse'
import { throttle } from 'lodash'

import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { reformatBotMessage } from '@sourcegraph/cody-shared/src/chat/viewHelpers'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Message as PromptMessage } from '@sourcegraph/cody-shared/src/sourcegraph-api'

import { intentDetector } from './services/sourcegraph-client'
import { streamCompletions } from './services/stream-completions'
import * as slackHelpers from './slack/helpers'
import { interactionFromMessage } from './slack/message-interaction'
import { SLACK_PREAMBLE } from './slack/preamble'

const IN_PROGRESS_MESSAGE = '...✍️'

/**
 * Handles human-generated messages in a Slack bot application.
 * Processes the messages, generates a prompt, and streams completions.
 */
export async function handleHumanMessage(event: AppMentionEvent, codebaseContext: CodebaseContext): Promise<void> {
    const channel = event.channel
    const thread_ts = slackHelpers.getEventTs(event)

    // Restore transcript from the Slack thread
    const messages = await slackHelpers.getThreadMessages(channel, thread_ts)
    const transcript = await restoreTranscriptFromSlackThread(codebaseContext, messages)

    // Send an in-progress message
    const response = await slackHelpers.postMessage(IN_PROGRESS_MESSAGE, channel, thread_ts)

    // Generate a prompt and start completion streaming
    const prompt = await transcript.toPrompt(SLACK_PREAMBLE)
    console.log('PROMPT', prompt)
    startCompletionStreaming(prompt, channel, response?.ts)
}

/**
 * Restores a transcript from the given Slack thread messages.
 */
async function restoreTranscriptFromSlackThread(codebaseContext: CodebaseContext, messages: SlackReplyMessage[]) {
    const transcript = new Transcript()
    const mergedMessages = mergeSequentialUserMessages(messages)

    for (const [index, message] of mergedMessages.entries()) {
        const interaction = await interactionFromMessage(
            message.human,
            intentDetector,
            // Fetch codebase context only for the last message
            index === mergedMessages.length - 1 ? codebaseContext : null
        )

        transcript.addInteraction(interaction)

        if (message.assistant?.text) {
            transcript.addAssistantResponse(message.assistant?.text)
        }
    }

    return transcript
}

/**
 * Starts streaming completions for the given prompt.
 */
function startCompletionStreaming(
    promptMessages: PromptMessage[],
    channel: string,
    inProgressMessageTs?: string
): void {
    streamCompletions(promptMessages, {
        onChange: text => {
            onBotMessageChange(reformatBotMessage(text, ''), channel, inProgressMessageTs)?.catch(console.error)
        },
        onComplete: () => {
            console.log('Streaming complete!')
        },
        onError: err => {
            console.error(err)
        },
    })
}

/**
 * Throttled function to update the bot message when there is a change.
 * Ensures message updates are throttled to avoid exceeding Slack API rate limits.
 */
const onBotMessageChange = throttle(async (text: string, channel, inProgressMessageTs?: string) => {
    if (inProgressMessageTs) {
        await slackHelpers.updateMessage(channel, inProgressMessageTs, text)
    } else {
        console.error('The in-progress mesasge is not found!')
    }
    // Throttle message updates to keep Slack API rate limiter happy.
}, 1000)

interface SlackInteraction {
    human: SlackReplyMessage
    assistant?: SlackReplyMessage
}

/**
 * Merges sequential user messages in a Slack thread to avoid missing important context.
 */
function mergeSequentialUserMessages(messages: SlackReplyMessage[]) {
    const mergedMessages: SlackInteraction[] = []

    for (const message of messages) {
        const text = message.text?.replace(/<@[\dA-Z]+>/gm, '').trim()
        const lastInteraction = mergedMessages[mergedMessages.length - 1]
        const updatedMessage = { ...message, blocks: undefined, text }

        if (!lastInteraction) {
            mergedMessages.push({ human: updatedMessage })
            continue
        }

        if (message.bot_id) {
            if (!lastInteraction.assistant) {
                lastInteraction.assistant = updatedMessage
            }
        } else if (!lastInteraction.assistant) {
            lastInteraction.human.text = `${lastInteraction.human.text || ''}; ${text}`
        } else {
            mergedMessages.push({ human: updatedMessage })
        }
    }

    return mergedMessages
}

import { AppMentionEvent } from '@slack/bolt'
import { Message as SlackReplyMessage } from '@slack/web-api/dist/response/ChannelsRepliesResponse'
import { throttle } from 'lodash'

import { Transcript } from '@sourcegraph/cody-shared/src/chat/transcript'
import { reformatBotMessage } from '@sourcegraph/cody-shared/src/chat/viewHelpers'
import { Message as PromptMessage } from '@sourcegraph/cody-shared/src/sourcegraph-api'

import { AppContext } from './constants'
import { streamCompletions } from './services/stream-completions'
import * as slackHelpers from './slack/helpers'
import { cleanupMessageForPrompt, getSlackInteraction } from './slack/message-interaction'
import { SLACK_PREAMBLE } from './slack/preamble'

const IN_PROGRESS_MESSAGE = '...✍️'

/**
 * Used to test Slack channel context fetching.
 * E.g., @cody-dev channel:ask-cody your prompt.
 */
function parseSlackChannelFilter(input: string): string | null {
    const match = input.match(/channel:([\w-]+)/)
    return match ? match[1] : null
}

/**
 * Handles human-generated messages in a Slack bot application.
 * Processes the messages, generates a prompt, and streams completions.
 */
export async function handleHumanMessage(event: AppMentionEvent, appContext: AppContext): Promise<void> {
    const channel = event.channel
    const thread_ts = slackHelpers.getEventTs(event)
    const slackChannelFilter = parseSlackChannelFilter(event.text)

    // Restore transcript from the Slack thread
    const [messages, channelName] = await Promise.all([
        slackHelpers.getThreadMessages(channel, thread_ts),
        slackHelpers.getSlackChannelName(channel),
    ])

    const transcript = await restoreTranscriptFromSlackThread(slackChannelFilter || channelName!, appContext, messages)

    // Send an in-progress message
    const response = await slackHelpers.postMessage(IN_PROGRESS_MESSAGE, channel, thread_ts)

    // Generate a prompt and start completion streaming
    const prompt = await transcript.toPrompt(SLACK_PREAMBLE)
    console.log('PROMPT', prompt)
    startCompletionStreaming(prompt, channel, transcript, response?.ts)
}

/**
 * Restores a transcript from the given Slack thread messages.
 */
async function restoreTranscriptFromSlackThread(
    channelName: string,
    appContext: AppContext,
    messages: SlackReplyMessage[]
) {
    const { codebaseContexts, vectorStore } = appContext
    const transcript = new Transcript()

    const mergedMessages = mergeSequentialUserMessages(messages)
    const newHumanMessage = mergedMessages.pop()!

    mergedMessages.forEach(message => {
        const slackInteraction = getSlackInteraction(message.human.text, message.assistant?.text)

        transcript.addInteraction(slackInteraction.getTranscriptInteraction())
    })

    const newHumanSlackInteraction = getSlackInteraction(newHumanMessage?.human.text)

    if (channelName === 'ask-cody') {
        await Promise.all([
            newHumanSlackInteraction.updateContextMessagesFromVectorStore(vectorStore, 3),
            newHumanSlackInteraction.updateContextMessages(codebaseContexts, 'github.com/sourcegraph/sourcegraph', {
                numCodeResults: 3,
                numTextResults: 5,
            }),
            newHumanSlackInteraction.updateContextMessages(codebaseContexts, 'github.com/sourcegraph/handbook', {
                numCodeResults: 0,
                numTextResults: 4,
            }),
        ])
    } else {
        await newHumanSlackInteraction.updateContextMessages(codebaseContexts, 'github.com/sourcegraph/sourcegraph', {
            numCodeResults: 12,
            numTextResults: 3,
        })
    }

    const lastInteraction = newHumanSlackInteraction.getTranscriptInteraction()
    transcript.addInteraction(lastInteraction)

    return transcript
}

/**
 * Starts streaming completions for the given prompt.
 */
function startCompletionStreaming(
    promptMessages: PromptMessage[],
    channel: string,
    transcript: Transcript,
    inProgressMessageTs?: string
): void {
    const lastInteraction = transcript.getLastInteraction()!

    const { contextFiles = [] } = lastInteraction.toChat().pop()!

    // Build the markdown list of file links.
    const contextFilesList = contextFiles
        .map(file => `[${file.fileName.split('/').pop()}](${file.fileName})`)
        .join(', ')

    const suffix = contextFiles.length > 0 ? '\n\n**Files used**:\n' + contextFilesList : ''

    streamCompletions(promptMessages, {
        onChange: text => {
            // console.log('Stream update: ', text)
            lastInteraction.setAssistantMessage({ ...lastInteraction.getAssistantMessage(), text })
            onBotMessageChange(channel, inProgressMessageTs, reformatBotMessage(text, '') + suffix)?.catch(
                console.error
            )
        },
        onComplete: () => {
            console.log('Streaming complete!', lastInteraction.getAssistantMessage().text)
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
const onBotMessageChange = throttle(async (channel, inProgressMessageTs: string | undefined, text: string) => {
    if (inProgressMessageTs) {
        await slackHelpers.updateMessage(channel, inProgressMessageTs, text)
    } else {
        console.error('The in-progress mesasge is not found!')
    }
    // Throttle message updates to keep Slack API rate limiter happy.
}, 1000)

interface SlackInteraction {
    human: { text: string }
    assistant?: { text: string }
}

/**
 * Merges sequential user messages in a Slack thread to avoid missing important context.
 */
function mergeSequentialUserMessages(messages: SlackReplyMessage[]) {
    const mergedMessages: SlackInteraction[] = []

    for (const message of messages) {
        const lastInteraction = mergedMessages[mergedMessages.length - 1]

        const text = cleanupMessageForPrompt(message.text || '', Boolean(message.bot_id))
        const updatedMessage = { text }

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

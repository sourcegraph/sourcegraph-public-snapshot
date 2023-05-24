import { KnownEventFromType, SlackEvent } from '@slack/bolt'
import { ChatPostMessageResponse } from '@slack/web-api'
import slackifyMarkdown from 'slackify-markdown'

import { webClient } from './init'

export async function getThreadMessages(channel: string, thread_ts: string) {
    const result = await webClient.conversations.replies({
        channel,
        ts: thread_ts,
    })

    return result?.messages || []
}

export async function getSlackChannelName(channel: string) {
    const result = await webClient.conversations.info({
        channel,
    })

    return result.channel?.name
}

export function getEventTs(event: KnownEventFromType<'app_mention'> | KnownEventFromType<'message'>) {
    if ('thread_ts' in event && event.thread_ts !== event.ts) {
        return event.thread_ts || event.ts
    }

    return event.ts
}

export const isBotEvent = (event: SlackEvent) => {
    return 'subtype' in event && event.subtype !== 'bot_message'
}

export async function updateMessage(channel: string, messageTs: string, newText: string): Promise<void> {
    const response = await webClient.chat.update({
        channel,
        ts: messageTs, // The timestamp of the message you want to update.
        text: slackifyMarkdown(newText), // The new text for the updated message.
        mrkdwn: true,
    })

    if (!response.ok) {
        throw new Error(`Error updating message: ${response.error}`)
    }
}

export async function updateMessageWithFileList(
    channel: string,
    messageTs: string,
    newText: string,
    fileList: string
): Promise<void> {
    const response = await webClient.chat.update({
        channel,
        blocks: [
            {
                type: 'section',
                text: {
                    type: 'mrkdwn',
                    text: slackifyMarkdown(fileList),
                },
            },
        ],
        ts: messageTs, // The timestamp of the message you want to update.
        text: slackifyMarkdown(newText), // The new text for the updated message.
        mrkdwn: true,
    })

    if (!response.ok) {
        throw new Error(`Error updating message: ${response.error}`)
    }
}

export async function postMessage(
    message: string,
    channel: string,
    thread_ts: string
): Promise<ChatPostMessageResponse | undefined> {
    const response = await webClient.chat.postMessage({
        channel,
        text: slackifyMarkdown(message),
        thread_ts, // Use the timestamp of the parent message to reply in the thread.
        mrkdwn: true,
    })

    if (!response.ok) {
        throw new Error(`Error sending message: ${response.error}`)
    }

    return response
}

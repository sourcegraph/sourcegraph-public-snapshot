import { CHARS_PER_TOKEN } from './limits'
import { parseEvents } from './parse'
import { CompletionCallbacks, Message } from './types'

export function humanInput(input: string): Message[] {
    return [
        {
            speaker: 'human',
            text: input,
        },
        { speaker: 'assistant', text: '' },
    ]
}

export const isError = (value: unknown): value is Error => value instanceof Error

function padTimePart(timePart: number): string {
    return timePart < 10 ? `0${timePart}` : timePart.toString()
}

export function getShortTimestamp(): string {
    const date = new Date()
    return `${padTimePart(date.getHours())}:${padTimePart(date.getMinutes())}`
}

export function truncateText(text: string, maxTokens: number): string {
    const maxLength = maxTokens * CHARS_PER_TOKEN
    return text.length <= maxLength ? text : text.slice(0, maxLength)
}

export function truncateTextStart(text: string, maxTokens: number): string {
    const maxLength = maxTokens * CHARS_PER_TOKEN
    return text.length <= maxLength ? text : text.slice(-maxLength - 1)
}

export const conversationStarter: Message[] = [
    {
        speaker: 'human',
        text: 'You are Cody, an AI-powered coding assistant created by Sourcegraph that performs the following actions:\n- Answer general programming questions\n- Answer questions about code that I have provided to you\n- Generate code that matches a written description\n- Explain what a section of code does\n\nIn your responses, you should obey the following rules:\n- Be as brief and concise as possible without losing clarity\n- Any code snippets should be markdown-formatted (placed in-between triple backticks like this "").\n- Answer questions only if you know the answer or can make a well-informed guess. Otherwise, tell me you don\'t know, and tell me what context I need to provide to you in order for you to answer the question.\n- Do not reference any file names or URLs, unless you are sure they exist.',
    },
    {
        speaker: 'assistant',
        text: 'Understood. I am Cody, an AI-powered coding assistant created by Sourcegraph and will follow the rules above',
    },
]

export const createRequestBody = (conversations: Message[]) => {
    return {
        messages: conversations,
        temperature: 0.2,
        maxTokensToSample: 1000,
        topK: -1,
        topP: -1,
    }
}

export function sendEvents(body: string, cb: CompletionCallbacks): void {
    const parseResult = parseEvents(body)
    if (isError(parseResult)) {
        console.error(parseResult)
        return
    }
    for (const event of parseResult.events) {
        switch (event.type) {
            case 'completion':
                cb.onChange(event.completion)
                break
            case 'error':
                cb.onError(event.error)
                break
            case 'done':
                cb.onComplete()
                break
        }
    }
}

// Check if user is an authorized user of the provided sg instance
export const isLoggedin = async (uri: string, token: string) => {
    if (!uri || !token) {
        console.error('incorrect uri:', uri)
        return false
    }
    const sgURL = new URL('/.api/graphql', uri).href
    const res = await fetch(sgURL, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `token ${token}`,
        },
        body: JSON.stringify({
            query: `
        query CurrentAuthState {
          currentUser {
              __typename
              id
              databaseID
              username
              avatarURL
              email
              displayName
          }
      }
      `,
        }),
    })
    chrome.storage.local.set({ sgCodyAuth: res.ok })
    return res.ok
}

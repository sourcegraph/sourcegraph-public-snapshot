import { Message, WSChatRequest, WSChatResponse } from '../types'

import { WSClient } from './wsclient'

interface ChatCallbacks {
    onChange: (text: string) => void
    onComplete: (text: string) => void
    onError: (message: string, originalErrorEvent?: ErrorEvent) => void
}

export class WSChatClient {
    public static async new(addr: string, accessToken: string): Promise<WSChatClient | null> {
        const wsclient = await WSClient.new<Omit<WSChatRequest, 'requestId'>, WSChatResponse>(addr, accessToken)
        if (!wsclient) {
            return null
        }
        return new WSChatClient(wsclient)
    }

    constructor(private wsclient: WSClient<Omit<WSChatRequest, 'requestId'>, WSChatResponse>) {}

    public chat(messages: Message[], callbacks: ChatCallbacks): Promise<() => void> {
        return this.wsclient.sendRequest(
            {
                kind: 'request',
                messages,
            },
            resp => {
                switch (resp.kind) {
                    case 'response:change':
                        callbacks.onChange(resp.message)
                        return false
                    case 'response:complete':
                        callbacks.onComplete(resp.message)
                        return true
                    case 'response:error':
                        callbacks.onError(resp.error)
                        return false
                    default:
                        return false
                }
            }
        )
    }
}

import { useMemo } from 'react'

import useWebSocket from 'react-use-websocket'
import { JsonValue, WebSocketHook } from 'react-use-websocket/dist/lib/types'

/**
 * For Sourcegraph team members only. For instructions, see
 * https://docs.google.com/document/d/1u7HYPmJFtDANtBgczzmAR0BmhM86drwDXCqx-F2jTEE/edit#.
 */
const CODY_ACCESS_TOKEN = localStorage.getItem('codyAccessToken')
const CODY_ENDPOINT_URL = localStorage.getItem('codyEndpointURL')

const getAuthenticatedEndpointURL = (endpointUrl: string, accessToken: string): string => {
    const url = new URL(endpointUrl)
    url.pathname = '/chat'
    url.searchParams.set('access_token', accessToken)
    return url.toString()
}

export function useCodyWebsocket(): Pick<
    WebSocketHook<JsonValue, MessageEvent<any> | null>,
    'sendMessage' | 'lastMessage' | 'readyState'
> {
    if (CODY_ENDPOINT_URL === null || CODY_ACCESS_TOKEN === null) {
        throw new Error('Cody is not configured')
    }
    const authenticatedEndpointURL = useMemo(
        () => getAuthenticatedEndpointURL(CODY_ENDPOINT_URL, CODY_ACCESS_TOKEN),
        []
    )

    const { sendMessage, lastMessage, readyState } = useWebSocket(authenticatedEndpointURL, {
        reconnectAttempts: 3,
        reconnectInterval: 500,
        shouldReconnect: () => true,
    })

    return { sendMessage, lastMessage, readyState }
}

export function getCodyCompletionOneShot(prompt: string): Promise<string> {
    if (CODY_ENDPOINT_URL === null || CODY_ACCESS_TOKEN === null) {
        throw new Error('Cody is not configured')
    }
    const authenticatedEndpointURL = getAuthenticatedEndpointURL(CODY_ENDPOINT_URL, CODY_ACCESS_TOKEN)
    const websocket = new WebSocket(authenticatedEndpointURL)
    return new Promise<string>((resolve, reject) => {
        websocket.addEventListener('open', () => {
            websocket.send(
                JSON.stringify({
                    requestId: 1,
                    messages: [{ speaker: 'you', text: prompt }],
                })
            )
        })
        websocket.addEventListener('message', event => {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
            const data = JSON.parse(event.data)
            if (data.kind === 'response:complete') {
                resolve((data.message as string).trim())
            }
        })
        websocket.addEventListener('error', () => {
            reject(new Error('websocket error'))
        })
    })
}

import React, { ChangeEventHandler, useCallback, useMemo, useState } from 'react'

import useWebSocket, { ReadyState } from 'react-use-websocket'

import { renderMarkdown } from '@sourcegraph/common'
import { Button, Form, H1, Input, Markdown, Text } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../cody/CodyIcon'

const CODY_ACCESS_TOKEN = localStorage.getItem('codyAccessToken')
const CODY_ENDPOINT_URL = localStorage.getItem('codyEndpointURL')

/**
 * For Sourcegraph team members only. For instructions, see
 * https://docs.google.com/document/d/1u7HYPmJFtDANtBgczzmAR0BmhM86drwDXCqx-F2jTEE/edit#.
 */
export const CodyPage: React.FunctionComponent<{}> = () => {
    if (CODY_ENDPOINT_URL === null || CODY_ACCESS_TOKEN === null) {
        throw new Error('Cody is not configured')
    }

    const authenticatedEndpointURL = useMemo(() => {
        const url = new URL(CODY_ENDPOINT_URL)
        url.pathname = '/chat'
        url.searchParams.set('access_token', CODY_ACCESS_TOKEN)
        return url
    }, [])
    const { sendMessage, lastMessage, readyState } = useWebSocket(authenticatedEndpointURL.toString())

    const [input, setInput] = useState('')
    const onInputChange = useCallback<ChangeEventHandler<HTMLInputElement>>(event => {
        setInput(event.currentTarget.value)
    }, [])

    const isReady = readyState === ReadyState.OPEN

    return (
        <div>
            <H1>
                <CodyIcon /> Cody
            </H1>
            <Form onSubmit={event => event.preventDefault()} className="d-flex">
                <Input type="text" onChange={onInputChange} value={input} disabled={!isReady} className="flex-1 mr-2" />
                <Button
                    type="submit"
                    variant="primary"
                    onClick={() =>
                        sendMessage(
                            JSON.stringify({
                                requestId: 1,
                                messages: [{ speaker: 'you', text: input }],
                            })
                        )
                    }
                    disabled={!isReady}
                    className="flex-0"
                >
                    Send
                </Button>
            </Form>
            <hr className="my-3" />
            {!isReady ? (
                <Text className="text-muted">Connecting to Cody...</Text>
            ) : (
                !lastMessage && <Text className="text-muted">Cody is ready.</Text>
            )}
            {lastMessage ? (
                <dl>
                    <dt>Cody:</dt>
                    <dd className="ml-3">
                        {/* eslint-disable-next-line @typescript-eslint/no-unsafe-member-access */}
                        <Markdown dangerousInnerHTML={renderMarkdown(JSON.parse(lastMessage.data).message)} />
                    </dd>
                </dl>
            ) : null}
        </div>
    )
}

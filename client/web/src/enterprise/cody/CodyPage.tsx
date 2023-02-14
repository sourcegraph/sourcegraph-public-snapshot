import { renderMarkdown } from '@sourcegraph/common'
import { Button, Form, Input, Markdown } from '@sourcegraph/wildcard'
import React, { ChangeEventHandler, useCallback, useMemo, useState } from 'react'
import useWebSocket, { ReadyState } from 'react-use-websocket'
import { CodyIcon } from './CodyIcon'

/**
 * Sourcegraph team members only: For instructions on how to set these, see
 * https://docs.google.com/document/d/1u7HYPmJFtDANtBgczzmAR0BmhM86drwDXCqx-F2jTEE/edit#.
 */
const CODY_ACCESS_TOKEN = localStorage.codyAccessToken
const CODY_ENDPOINT_URL = localStorage.codyEndpointURL

/**
 * For Sourcegraph team members only. For instructions, see
 * https://docs.google.com/document/d/1u7HYPmJFtDANtBgczzmAR0BmhM86drwDXCqx-F2jTEE/edit#.
 */
export const CodyPage: React.FunctionComponent<{}> = () => {
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
            <h1>
                <CodyIcon className="icon-inline" /> Cody
            </h1>
            <Form onSubmit={event => event.preventDefault()} className="d-flex">
                <Input type="text" onChange={onInputChange} value={input} className="flex-1 mr-2" />
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
            {lastMessage ? (
                <dl>
                    <dt>Cody:</dt>
                    <dd className="ml-3">
                        <Markdown dangerousInnerHTML={renderMarkdown(JSON.parse(lastMessage.data).message)} />
                    </dd>
                </dl>
            ) : null}
        </div>
    )
}

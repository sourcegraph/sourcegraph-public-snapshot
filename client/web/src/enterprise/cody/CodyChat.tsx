import React, { ChangeEventHandler, useCallback, useMemo, useState } from 'react'

import useWebSocket, { ReadyState } from 'react-use-websocket'

import { renderMarkdown } from '@sourcegraph/common'
import { Button, Form, Input, Markdown, Text } from '@sourcegraph/wildcard'

import { useCodyWebsocket } from './api'

export const CodyChat: React.FunctionComponent<{ promptPrefix?: string }> = ({ promptPrefix }) => {
    const { sendMessage, lastMessage, readyState } = useCodyWebsocket()

    const [input, setInput] = useState('')
    const onInputChange = useCallback<ChangeEventHandler<HTMLInputElement>>(event => {
        setInput(event.currentTarget.value)
    }, [])

    const isReady = readyState === ReadyState.OPEN

    return (
        <div>
            <Form onSubmit={event => event.preventDefault()} className="d-flex">
                <Input
                    type="text"
                    onChange={onInputChange}
                    value={input}
                    disabled={!isReady}
                    className="flex-1 mr-2"
                    autoFocus={true}
                />
                <Button
                    type="submit"
                    variant="primary"
                    onClick={() =>
                        sendMessage(
                            JSON.stringify({
                                requestId: 1,
                                messages: [
                                    {
                                        speaker: 'you',
                                        text: `${promptPrefix}\n\n${input}\n\nHuman: If the response contains code, format only the code in Markdown code blocks. Do not wrap prose in Markdown code blocks or backticks. If you are not certain of the answer, do not guess.\n\nAssistant: `,
                                    },
                                ],
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

import React, { useCallback, useEffect, useState } from 'react'

import { Client, createClient } from '@sourcegraph/cody-shared/src/chat/client'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { CodySvg } from '@sourcegraph/cody-ui/src/utils/icons'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { Alert, LoadingSpinner } from '@sourcegraph/wildcard'

import { Chat } from './Chat'
import { Settings } from './settings/Settings'
import { useConfig } from './settings/useConfig'

import styles from './App.module.css'

export const App: React.FunctionComponent = () => {
    const [config, setConfig] = useConfig()
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])

    const [client, setClient] = useState<Client | ErrorLike>()
    useEffect(() => {
        setMessageInProgress(null)
        setTranscript([])
        createClient({ config, accessToken: config.accessToken, setMessageInProgress, setTranscript }).then(
            setClient,
            setClient
        )
    }, [config])

    const onSubmit = useCallback(
        (text: string) => {
            if (client && !isErrorLike(client)) {
                client.submitMessage(text)
            }
        },
        [client]
    )

    return (
        <div className={styles.container}>
            <header className={styles.header}>
                <h1>
                    <CodySvg /> Cody
                </h1>
                <Settings config={config} setConfig={setConfig} />
            </header>
            <main className={styles.main}>
                {!client ? (
                    <>
                        <LoadingSpinner />
                    </>
                ) : isErrorLike(client) ? (
                    <Alert className={styles.alert} variant="danger">
                        {client.message}
                    </Alert>
                ) : (
                    <>
                        <Chat
                            messageInProgress={messageInProgress}
                            transcript={transcript}
                            contextStatus={{ codebase: config.codebase }}
                            formInput={formInput}
                            setFormInput={setFormInput}
                            inputHistory={inputHistory}
                            setInputHistory={setInputHistory}
                            onSubmit={onSubmit}
                        />
                    </>
                )}
            </main>
        </div>
    )
}

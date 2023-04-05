import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiSend } from '@mdi/js'

import { Client, createClient, ClientInit } from '@sourcegraph/cody-shared/src/chat/client'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Chat, ChatUISubmitButtonProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { Terms } from '@sourcegraph/cody-ui/src/Terms'
import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Icon } from '@sourcegraph/wildcard'

import { Scalars } from '../../../graphql-operations'

import styles from './RepoRevisionSidebarCody.module.scss'

export const RepoRevisionSidebarCody: React.FunctionComponent<
    {
        repoName: string
        repoID: Scalars['ID']

        /** The path of the file or directory currently shown in the content area */
        activePath: string

        focusKey?: string
    } & Partial<RevisionSpec>
> = ({ repoName }) => {
    const config = useMemo<ClientInit['config']>(
        () => ({
            serverEndpoint: window.location.origin,
            useContext: 'embeddings',
            codebase: repoName,
        }),
        [repoName]
    )
    const [messageInProgress, setMessageInProgress] = useState<ChatMessage | null>(null)
    const [transcript, setTranscript] = useState<ChatMessage[]>([])
    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])

    const [client, setClient] = useState<Client | ErrorLike>()
    useEffect(() => {
        setMessageInProgress(null)
        setTranscript([])
        createClient({ config, accessToken: null, setMessageInProgress, setTranscript }).then(setClient, setClient)
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
        <Chat
            messageInProgress={messageInProgress}
            transcript={transcript}
            formInput={formInput}
            setFormInput={setFormInput}
            inputHistory={inputHistory}
            setInputHistory={setInputHistory}
            onSubmit={onSubmit}
            submitButtonComponent={SubmitButton}
            fileLinkComponent={FileLink}
            className={styles.container}
            afterTips={
                <details className="small mt-2">
                    <summary>Terms</summary>
                    <Terms />
                </details>
            }
            bubbleContentClassName={styles.bubbleContent}
            humanBubbleContentClassName={styles.humanBubbleContent}
            botBubbleContentClassName={styles.botBubbleContent}
            bubbleFooterClassName="text-muted small"
            bubbleLoaderDotClassName={styles.bubbleLoaderDot}
            inputRowClassName={styles.inputRow}
            chatInputClassName={styles.chatInput}
        />
    )
}

const SubmitButton: React.FunctionComponent<ChatUISubmitButtonProps> = ({ className, disabled, onClick }) => (
    <button className={className} type="submit" disabled={disabled} onClick={onClick}>
        <Icon svgPath={mdiSend} />
    </button>
)

const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => <>{path}</>

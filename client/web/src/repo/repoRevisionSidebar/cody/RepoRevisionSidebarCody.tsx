import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiSend } from '@mdi/js'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Chat, ChatUISubmitButtonProps } from '@sourcegraph/cody-ui/src/Chat'
import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'
import { Terms } from '@sourcegraph/cody-ui/src/Terms'
import { RevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Icon } from '@sourcegraph/wildcard'

import { Scalars } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import styles from './RepoRevisionSidebarCody.module.scss'

export const RepoRevisionSidebarCody: React.FunctionComponent<
    {
        repoName: string
        repoID: Scalars['ID']
        activePath: string
        focusKey?: string
        onSubmit: (text: string) => void
        messageInProgress: ChatMessage | null
        transcript: ChatMessage[]
    } & Partial<RevisionSpec>
> = ({ repoName, activePath, onSubmit, messageInProgress, transcript }) => {
    useEffect(() => {
        eventLogger.log('web:codySidebar:view', { repo: repoName })
    }, [repoName])

    const [formInput, setFormInput] = useState('')
    const [inputHistory, setInputHistory] = useState<string[] | []>([])

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

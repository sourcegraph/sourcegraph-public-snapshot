import React, { useEffect, useRef } from 'react'

import classNames from 'classnames'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { FileLinkProps } from './ContextFiles'
import { TranscriptItem, TranscriptItemClassNames } from './TranscriptItem'

import styles from './Transcript.module.css'

export const Transcript: React.FunctionComponent<
    {
        transcript: ChatMessage[]
        messageInProgress: ChatMessage | null
        fileLinkComponent: React.FunctionComponent<FileLinkProps>
        className?: string
    } & TranscriptItemClassNames
> = ({
    transcript,
    messageInProgress,
    fileLinkComponent,
    className,
    codeBlocksCopyButtonClassName,
    transcriptItemClassName,
    humanTranscriptItemClassName,
    transcriptItemParticipantClassName,
    transcriptActionClassName,
}) => {
    const transcriptContainerRef = useRef<HTMLDivElement>(null)
    useEffect(() => {
        if (transcriptContainerRef.current) {
            // Only scroll if the user didn't scroll up manually more than the scrolling threshold.
            // That is so that you can freely copy content or read up on older content while new
            // content is being produced.
            //
            // We allow some small threshold for "what is considered not scrolled up" so that
            // minimal scroll doesn't affect it (ie. if I'm not all the way scrolled down by like a
            // pixel or two, I probably still want it to scroll).
            const SCROLL_THRESHOLD = 100
            const delta = Math.abs(
                transcriptContainerRef.current.scrollHeight -
                    transcriptContainerRef.current.offsetHeight -
                    transcriptContainerRef.current.scrollTop
            )
            if (delta < SCROLL_THRESHOLD) {
                transcriptContainerRef.current.scrollTo({
                    top: transcriptContainerRef.current.scrollHeight,
                })
            }
        }
    }, [transcript, transcriptContainerRef])

    // Scroll down whenever a new message is received.
    const lastMessageSpeaker = transcript[transcript.length - 1]?.speaker
    useEffect(() => {
        transcriptContainerRef.current?.scrollTo({
            top: transcriptContainerRef.current.scrollHeight,
        })
    }, [lastMessageSpeaker])

    return (
        <div ref={transcriptContainerRef} className={classNames(className, styles.container)}>
            {transcript.map((message, index) => (
                <TranscriptItem
                    // eslint-disable-next-line react/no-array-index-key
                    key={index}
                    message={message}
                    inProgress={false}
                    fileLinkComponent={fileLinkComponent}
                    codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                    transcriptItemClassName={transcriptItemClassName}
                    humanTranscriptItemClassName={humanTranscriptItemClassName}
                    transcriptItemParticipantClassName={transcriptItemParticipantClassName}
                    transcriptActionClassName={transcriptActionClassName}
                />
            ))}
            {messageInProgress && messageInProgress.speaker === 'assistant' && (
                <TranscriptItem
                    message={messageInProgress}
                    inProgress={true}
                    fileLinkComponent={fileLinkComponent}
                    codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                    transcriptItemClassName={transcriptItemClassName}
                    transcriptItemParticipantClassName={transcriptItemParticipantClassName}
                    transcriptActionClassName={transcriptActionClassName}
                />
            )}
        </div>
    )
}

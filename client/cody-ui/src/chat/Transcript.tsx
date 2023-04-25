import React, { useEffect, useRef } from 'react'

import classNames from 'classnames'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { ChatUITextAreaProps, EditButtonProps, FeedbackButtonsProps, CopyButtonProps } from '../Chat'

import { FileLinkProps } from './ContextFiles'
import { TranscriptItem, TranscriptItemClassNames } from './TranscriptItem'

import styles from './Transcript.module.css'

export const Transcript: React.FunctionComponent<
    {
        transcript: ChatMessage[]
        messageInProgress: ChatMessage | null
        messageBeingEdited: boolean
        setMessageBeingEdited: (input: boolean) => void
        fileLinkComponent: React.FunctionComponent<FileLinkProps>
        className?: string
        textAreaComponent?: React.FunctionComponent<ChatUITextAreaProps>
        EditButtonContainer?: React.FunctionComponent<EditButtonProps>
        editButtonOnSubmit?: (text: string) => void
        FeedbackButtonsContainer?: React.FunctionComponent<FeedbackButtonsProps>
        feedbackButtonsOnSubmit?: (text: string) => void
        copyButtonOnSubmit?: CopyButtonProps['copyButtonOnSubmit']
    } & TranscriptItemClassNames
> = ({
    transcript,
    messageInProgress,
    messageBeingEdited,
    setMessageBeingEdited,
    fileLinkComponent,
    className,
    codeBlocksCopyButtonClassName,
    transcriptItemClassName,
    humanTranscriptItemClassName,
    transcriptItemParticipantClassName,
    transcriptActionClassName,
    textAreaComponent,
    EditButtonContainer,
    editButtonOnSubmit,
    FeedbackButtonsContainer,
    feedbackButtonsOnSubmit,
    copyButtonOnSubmit,
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
            {transcript.map(
                (message, index) =>
                    message?.displayText &&
                    (!messageInProgress || index !== transcript.length - 1) && (
                        <TranscriptItem
                            // eslint-disable-next-line react/no-array-index-key
                            key={index}
                            message={message}
                            inProgress={false}
                            beingEdited={index > 0 && transcript.length - index === 2 && messageBeingEdited}
                            setBeingEdited={setMessageBeingEdited}
                            fileLinkComponent={fileLinkComponent}
                            codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                            transcriptItemClassName={transcriptItemClassName}
                            humanTranscriptItemClassName={humanTranscriptItemClassName}
                            transcriptItemParticipantClassName={transcriptItemParticipantClassName}
                            transcriptActionClassName={transcriptActionClassName}
                            textAreaComponent={textAreaComponent}
                            EditButtonContainer={EditButtonContainer}
                            editButtonOnSubmit={editButtonOnSubmit}
                            showEditButton={index > 0 && transcript.length - index === 2}
                            FeedbackButtonsContainer={FeedbackButtonsContainer}
                            feedbackButtonsOnSubmit={feedbackButtonsOnSubmit}
                            copyButtonOnSubmit={copyButtonOnSubmit}
                            showFeedbackButtons={index > 0 && transcript.length - index === 1}
                        />
                    )
            )}
            {messageInProgress && messageInProgress.speaker === 'assistant' && (
                <TranscriptItem
                    message={messageInProgress}
                    inProgress={true}
                    beingEdited={false}
                    setBeingEdited={setMessageBeingEdited}
                    fileLinkComponent={fileLinkComponent}
                    codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                    transcriptItemClassName={transcriptItemClassName}
                    transcriptItemParticipantClassName={transcriptItemParticipantClassName}
                    transcriptActionClassName={transcriptActionClassName}
                    showEditButton={false}
                    showFeedbackButtons={false}
                    copyButtonOnSubmit={copyButtonOnSubmit}
                />
            )}
        </div>
    )
}

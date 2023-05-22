import React, { useState } from 'react'

import classNames from 'classnames'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import {
    ChatUITextAreaProps,
    EditButtonProps,
    FeedbackButtonsProps,
    CopyButtonProps,
    ChatUISubmitButtonProps,
} from '../Chat'
import { CodySvg } from '../utils/icons'

import { BlinkingCursor } from './BlinkingCursor'
import { CodeBlocks } from './CodeBlocks'
import { ContextFiles, FileLinkProps } from './ContextFiles'

import styles from './TranscriptItem.module.css'

/**
 * CSS class names used for the {@link TranscriptItem} component.
 */
export interface TranscriptItemClassNames {
    transcriptItemClassName?: string
    humanTranscriptItemClassName?: string
    transcriptItemParticipantClassName?: string
    codeBlocksCopyButtonClassName?: string
    transcriptActionClassName?: string
    chatInputClassName?: string
}

/**
 * A single message in the chat trans cript.
 */
export const TranscriptItem: React.FunctionComponent<
    {
        message: ChatMessage
        inProgress: boolean
        beingEdited: boolean
        setBeingEdited: (input: boolean) => void
        fileLinkComponent: React.FunctionComponent<FileLinkProps>
        textAreaComponent?: React.FunctionComponent<ChatUITextAreaProps>
        EditButtonContainer?: React.FunctionComponent<EditButtonProps>
        editButtonOnSubmit?: (text: string) => void
        showEditButton: boolean
        FeedbackButtonsContainer?: React.FunctionComponent<FeedbackButtonsProps>
        feedbackButtonsOnSubmit?: (text: string) => void
        showFeedbackButtons: boolean
        copyButtonOnSubmit?: CopyButtonProps['copyButtonOnSubmit']
        submitButtonComponent?: React.FunctionComponent<ChatUISubmitButtonProps>
    } & TranscriptItemClassNames
> = ({
    message,
    inProgress,
    beingEdited,
    setBeingEdited,
    fileLinkComponent,
    transcriptItemClassName,
    humanTranscriptItemClassName,
    transcriptItemParticipantClassName,
    codeBlocksCopyButtonClassName,
    transcriptActionClassName,
    textAreaComponent: TextArea,
    EditButtonContainer,
    editButtonOnSubmit,
    showEditButton,
    FeedbackButtonsContainer,
    feedbackButtonsOnSubmit,
    showFeedbackButtons,
    copyButtonOnSubmit,
    submitButtonComponent: SubmitButton,
    chatInputClassName,
}) => {
    const [formInput, setFormInput] = useState<string>(message.displayText ?? '')
    const textarea =
        TextArea && beingEdited && editButtonOnSubmit && SubmitButton ? (
            <div className={styles.textAreaContainer}>
                <TextArea
                    className={classNames(styles.chatInput, chatInputClassName)}
                    rows={5}
                    value={formInput}
                    autoFocus={true}
                    required={true}
                    onInput={event => setFormInput((event.target as HTMLInputElement).value)}
                    onKeyDown={event => {
                        if (event.key === 'Escape') {
                            setBeingEdited(false)
                        }

                        if (
                            event.key === 'Enter' &&
                            !event.shiftKey &&
                            !event.nativeEvent.isComposing &&
                            formInput.trim()
                        ) {
                            event.preventDefault()
                            setBeingEdited(false)
                            editButtonOnSubmit(formInput)
                        }
                    }}
                />
                <SubmitButton
                    className={styles.submitButton}
                    onClick={() => {
                        setBeingEdited(false)
                        editButtonOnSubmit(formInput)
                    }}
                    disabled={formInput.length === 0}
                />
            </div>
        ) : null

    return (
        <div
            className={classNames(
                styles.row,
                transcriptItemClassName,
                message.speaker === 'human' ? humanTranscriptItemClassName : null
            )}
        >
            <header className={classNames(styles.participant, transcriptItemParticipantClassName)}>
                <h2 className={styles.participantName}>
                    {message.speaker === 'assistant' ? (
                        <>
                            <CodySvg className={styles.participantAvatar} /> Cody
                        </>
                    ) : (
                        'Me'
                    )}
                </h2>
                {/* display edit buttons on last user message, feedback buttons on last assistant message only */}
                <div className={styles.participantName}>
                    {showEditButton &&
                        EditButtonContainer &&
                        editButtonOnSubmit &&
                        TextArea &&
                        message.speaker === 'human' && (
                            <EditButtonContainer
                                className={styles.FeedbackEditButtonsContainer}
                                messageBeingEdited={beingEdited}
                                setMessageBeingEdited={setBeingEdited}
                            />
                        )}
                    {showFeedbackButtons &&
                        FeedbackButtonsContainer &&
                        feedbackButtonsOnSubmit &&
                        message.speaker === 'assistant' && (
                            <FeedbackButtonsContainer
                                className={styles.FeedbackEditButtonsContainer}
                                feedbackButtonsOnSubmit={feedbackButtonsOnSubmit}
                            />
                        )}
                </div>
            </header>
            {message.contextFiles && message.contextFiles.length > 0 && (
                <div className={styles.actions}>
                    <ContextFiles
                        contextFiles={message.contextFiles}
                        fileLinkComponent={fileLinkComponent}
                        className={transcriptActionClassName}
                    />
                </div>
            )}
            <div className={classNames(styles.contentPadding, textarea ? undefined : styles.content)}>
                {message.displayText ? (
                    textarea ?? (
                        <CodeBlocks
                            displayText={message.displayText}
                            copyButtonClassName={codeBlocksCopyButtonClassName}
                            CopyButtonProps={copyButtonOnSubmit}
                        />
                    )
                ) : inProgress ? (
                    <BlinkingCursor />
                ) : null}
            </div>
        </div>
    )
}

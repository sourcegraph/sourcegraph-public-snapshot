import React from 'react'

import classNames from 'classnames'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { FeedbackButtonsProps } from '../Chat'
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
}

/**
 * A single message in the chat trans cript.
 */
export const TranscriptItem: React.FunctionComponent<
    {
        message: ChatMessage
        inProgress: boolean
        fileLinkComponent: React.FunctionComponent<FileLinkProps>
        FeedbackButtonsContainer?: React.FunctionComponent<FeedbackButtonsProps>
        feedbackButtonsOnSubmit?: (text: string) => void
        showFeedbackButtons: boolean
    } & TranscriptItemClassNames
> = ({
    message,
    inProgress,
    fileLinkComponent,
    transcriptItemClassName,
    humanTranscriptItemClassName,
    transcriptItemParticipantClassName,
    codeBlocksCopyButtonClassName,
    transcriptActionClassName,
    FeedbackButtonsContainer,
    feedbackButtonsOnSubmit,
    showFeedbackButtons,
}) => (
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
            {/* display feedback buttons on last assistant message only */}
            <div className={styles.participantName}>
                {showFeedbackButtons &&
                    FeedbackButtonsContainer &&
                    feedbackButtonsOnSubmit &&
                    message.speaker === 'assistant' && (
                        <FeedbackButtonsContainer
                            className={styles.FeedbackButtonsContainer}
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
        <div className={classNames(styles.content)}>
            {message.displayText ? (
                <CodeBlocks displayText={message.displayText} copyButtonClassName={codeBlocksCopyButtonClassName} />
            ) : inProgress ? (
                <span>
                    Fetching context... <BlinkingCursor />
                </span>
            ) : null}
        </div>
    </div>
)

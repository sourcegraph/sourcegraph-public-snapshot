import React from 'react'

import classNames from 'classnames'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { ChatMessageLoading } from './ChatMessageLoading'
import { CodeBlocks } from './CodeBlocks'
import { ContextFiles, FileLinkProps } from './ContextFiles'

import styles from './ChatMessageRow.module.css'

export interface ChatMessageRowClassNames {
    bubbleContentClassName?: string
    bubbleClassName?: string
    bubbleRowClassName?: string
    humanBubbleContentClassName?: string
    botBubbleContentClassName?: string
    codeBlocksCopyButtonClassName?: string
    bubbleFooterClassName?: string
    bubbleLoaderDotClassName?: string
}

export const ChatMessageRow: React.FunctionComponent<
    {
        message: ChatMessage
        inProgress: boolean
        fileLinkComponent: React.FunctionComponent<FileLinkProps>
        className?: string
    } & ChatMessageRowClassNames
> = ({
    message,
    inProgress,
    fileLinkComponent,
    className,
    bubbleContentClassName,
    bubbleClassName,
    bubbleRowClassName,
    humanBubbleContentClassName,
    botBubbleContentClassName,
    codeBlocksCopyButtonClassName,
    bubbleFooterClassName,
    bubbleLoaderDotClassName,
}) => {
    const classNamePrefix = message.speaker === 'human' ? 'human' : 'bot'
    return (
        <div
            className={classNames(
                className,
                styles.bubbleRow,
                bubbleRowClassName,
                styles[`${classNamePrefix}BubbleRow`]
            )}
        >
            <div className={classNames(styles.bubble, bubbleClassName)}>
                <div
                    className={classNames(
                        styles.bubbleContent,
                        styles[`${classNamePrefix}BubbleContent`],
                        bubbleContentClassName,
                        message.speaker === 'human' ? humanBubbleContentClassName : botBubbleContentClassName
                    )}
                >
                    {message.displayText ? (
                        <>
                            <CodeBlocks
                                displayText={message.displayText}
                                copyButtonClassName={codeBlocksCopyButtonClassName}
                            />
                            {message.contextFiles && message.contextFiles.length > 0 && (
                                <ContextFiles
                                    contextFiles={message.contextFiles}
                                    fileLinkComponent={fileLinkComponent}
                                />
                            )}
                        </>
                    ) : inProgress ? (
                        <ChatMessageLoading bubbleLoaderDotClassName={bubbleLoaderDotClassName} />
                    ) : null}
                </div>
                <div
                    className={classNames(
                        styles.bubbleFooter,
                        styles[`${classNamePrefix}BubbleFooter`],
                        bubbleFooterClassName
                    )}
                >
                    {inProgress ? (
                        <span>Cody is typing...</span>
                    ) : (
                        <div className={styles.bubbleFooterTimestamp}>{`${
                            message.speaker === 'assistant' ? 'Cody' : 'Me'
                        } · ${message.timestamp}`}</div>
                    )}
                </div>
            </div>
        </div>
    )
}

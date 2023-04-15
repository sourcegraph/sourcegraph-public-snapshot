import React from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { CodeBlocks } from './CodeBlocks'
import { ContextFiles, FileLinkProps } from './ContextFiles'

import styles from './ChatMessages.module.css'

export interface ChatMessagesClassNames {
    bubbleContentClassName?: string
    bubbleClassName?: string
    bubbleRowClassName?: string
    humanBubbleContentClassName?: string
    botBubbleContentClassName?: string
    codeBlocksCopyButtonClassName?: string
    bubbleFooterClassName?: string
    bubbleLoaderDotClassName?: string
}

export const ChatMessages: React.FunctionComponent<
    {
        messageInProgress: ChatMessage | null
        transcript: ChatMessage[]
        fileLinkComponent: React.FunctionComponent<FileLinkProps>
        className?: string
    } & ChatMessagesClassNames
> = ({
    messageInProgress,
    transcript,
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
    const getBubbleClassName = (speaker: string): string => (speaker === 'human' ? 'human' : 'bot')

    return (
        <div className={className}>
            {transcript.map((message, index) => (
                <div
                    // eslint-disable-next-line react/no-array-index-key
                    key={`message-${index}`}
                    className={classNames(
                        styles.bubbleRow,
                        bubbleRowClassName,
                        styles[`${getBubbleClassName(message.speaker)}BubbleRow`]
                    )}
                >
                    <div className={classNames(styles.bubble, bubbleClassName)}>
                        <div
                            className={classNames(
                                styles.bubbleContent,
                                styles[`${getBubbleClassName(message.speaker)}BubbleContent`],
                                bubbleContentClassName,
                                message.speaker === 'human' ? humanBubbleContentClassName : botBubbleContentClassName
                            )}
                        >
                            {message.displayText && (
                                <CodeBlocks
                                    displayText={message.displayText}
                                    copyButtonClassName={codeBlocksCopyButtonClassName}
                                />
                            )}
                            {message.contextFiles && message.contextFiles.length > 0 && (
                                <ContextFiles
                                    contextFiles={message.contextFiles}
                                    fileLinkComponent={fileLinkComponent}
                                />
                            )}
                        </div>
                        <div
                            className={classNames(
                                styles.bubbleFooter,
                                styles[`${getBubbleClassName(message.speaker)}BubbleFooter`],
                                bubbleFooterClassName
                            )}
                        >
                            <div className={styles.bubbleFooterTimestamp}>{`${
                                message.speaker === 'assistant' ? 'Cody' : 'Me'
                            } · ${message.timestamp}`}</div>
                        </div>
                    </div>
                </div>
            ))}

            {messageInProgress && messageInProgress.speaker === 'assistant' && (
                <div className={classNames(styles.bubbleRow, styles.botBubbleRow)}>
                    <div className={styles.bubble}>
                        <div
                            className={classNames(
                                styles.bubbleContent,
                                styles.botBubbleContent,
                                bubbleContentClassName,
                                botBubbleContentClassName
                            )}
                        >
                            {messageInProgress.displayText ? (
                                <p
                                    dangerouslySetInnerHTML={{
                                        __html: renderMarkdown(messageInProgress.displayText),
                                    }}
                                />
                            ) : (
                                <div className={styles.bubbleLoader}>
                                    <div className={classNames(styles.bubbleLoaderDot, bubbleLoaderDotClassName)} />
                                    <div className={classNames(styles.bubbleLoaderDot, bubbleLoaderDotClassName)} />
                                    <div className={classNames(styles.bubbleLoaderDot, bubbleLoaderDotClassName)} />
                                </div>
                            )}
                        </div>
                        <div className={styles.bubbleFooter}>
                            <span>Cody is typing...</span>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}

import React from 'react'

import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { ChatMessageRow, ChatMessageRowClassNames } from './ChatMessageRow'
import { FileLinkProps } from './ContextFiles'

export interface ChatMessagesClassNames extends ChatMessageRowClassNames {}

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
}) => (
    <div className={className}>
        {transcript.map((message, index) => (
            <ChatMessageRow
                // eslint-disable-next-line react/no-array-index-key
                key={index}
                message={message}
                inProgress={false}
                fileLinkComponent={fileLinkComponent}
                bubbleContentClassName={bubbleContentClassName}
                bubbleClassName={bubbleClassName}
                bubbleRowClassName={bubbleRowClassName}
                humanBubbleContentClassName={humanBubbleContentClassName}
                botBubbleContentClassName={botBubbleContentClassName}
                codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                bubbleFooterClassName={bubbleFooterClassName}
                bubbleLoaderDotClassName={bubbleLoaderDotClassName}
            />
        ))}
        {messageInProgress && messageInProgress.speaker === 'assistant' && (
            <ChatMessageRow
                message={messageInProgress}
                inProgress={true}
                fileLinkComponent={fileLinkComponent}
                bubbleContentClassName={bubbleContentClassName}
                bubbleClassName={bubbleClassName}
                bubbleRowClassName={bubbleRowClassName}
                humanBubbleContentClassName={humanBubbleContentClassName}
                botBubbleContentClassName={botBubbleContentClassName}
                codeBlocksCopyButtonClassName={codeBlocksCopyButtonClassName}
                bubbleFooterClassName={bubbleFooterClassName}
                bubbleLoaderDotClassName={bubbleLoaderDotClassName}
            />
        )}
    </div>
)

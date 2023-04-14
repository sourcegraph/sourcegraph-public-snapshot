import React, { useCallback, useEffect, useRef, useState } from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { CodeBlocks } from './chat/CodeBlocks'
import { ContextFiles, FileLinkProps } from './chat/ContextFiles'
import { Tips } from './Tips'

import styles from './Chat.module.css'

const SCROLL_THRESHOLD = 15
interface ChatProps extends ChatClassNames {
    messageInProgress: ChatMessage | null
    transcript: ChatMessage[]
    formInput: string
    setFormInput: (input: string) => void
    inputHistory: string[]
    setInputHistory: (history: string[]) => void
    onSubmit: (text: string) => void
    textAreaComponent: React.FunctionComponent<ChatUITextAreaProps>
    submitButtonComponent: React.FunctionComponent<ChatUISubmitButtonProps>
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
    tipsRecommendations?: JSX.Element[]
    afterTips?: JSX.Element
    className?: string
}

interface ChatClassNames {
    transcriptContainerClassName?: string
    bubbleContentClassName?: string
    bubbleClassName?: string
    bubbleRowClassName?: string
    humanBubbleContentClassName?: string
    botBubbleContentClassName?: string
    codeBlocksCopyButtonClassName?: string
    bubbleFooterClassName?: string
    bubbleLoaderDotClassName?: string
    inputRowClassName?: string
    chatInputClassName?: string
}

export interface ChatUITextAreaProps {
    className: string
    rows: number
    autoFocus: boolean
    value: string
    required: boolean
    onInput: React.FormEventHandler<HTMLElement>
    onKeyDown: React.KeyboardEventHandler<HTMLElement>
}

export interface ChatUISubmitButtonProps {
    className: string
    disabled: boolean
    onClick: (event: React.MouseEvent<HTMLButtonElement>) => void
}
export const Chat: React.FunctionComponent<ChatProps> = ({
    messageInProgress,
    transcript,
    formInput,
    setFormInput,
    inputHistory,
    setInputHistory,
    onSubmit,
    textAreaComponent: TextArea,
    submitButtonComponent: SubmitButton,
    fileLinkComponent,
    tipsRecommendations,
    afterTips,
    className,
    transcriptContainerClassName,
    bubbleContentClassName,
    bubbleClassName,
    bubbleRowClassName,
    humanBubbleContentClassName,
    botBubbleContentClassName,
    codeBlocksCopyButtonClassName,
    bubbleFooterClassName,
    bubbleLoaderDotClassName,
    inputRowClassName,
    chatInputClassName,
}) => {
    const [inputRows, setInputRows] = useState(5)
    const [historyIndex, setHistoryIndex] = useState(inputHistory.length)
    const transcriptContainerRef = useRef<HTMLDivElement>(null)

    const inputHandler = useCallback(
        (inputValue: string): void => {
            const rowsCount = inputValue.match(/\n/g)?.length
            if (rowsCount) {
                setInputRows(rowsCount < 5 ? 5 : rowsCount > 25 ? 25 : rowsCount)
            } else {
                setInputRows(5)
            }
            setFormInput(inputValue)
            if (inputValue !== inputHistory[historyIndex]) {
                setHistoryIndex(inputHistory.length)
            }
        },
        [historyIndex, inputHistory, setFormInput]
    )

    const onChatSubmit = useCallback((): void => {
        // Submit chat only when input is not empty
        if (formInput !== undefined) {
            onSubmit(formInput)
            setHistoryIndex(inputHistory.length + 1)
            setInputHistory([...inputHistory, formInput])
            setInputRows(5)
            setFormInput('')
        }
    }, [formInput, inputHistory, onSubmit, setFormInput, setInputHistory])

    const onChatKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>): void => {
            // Submit input on Enter press (without shift) and
            // trim the formInput to make sure input value is not empty.
            if (
                event.key === 'Enter' &&
                !event.shiftKey &&
                !event.nativeEvent.isComposing &&
                formInput &&
                formInput.trim()
            ) {
                event.preventDefault()
                event.stopPropagation()
                onChatSubmit()
            }
            // Loop through input history on up arrow press
            if (event.key === 'ArrowUp' && inputHistory.length) {
                if (formInput === inputHistory[historyIndex] || !formInput) {
                    const newIndex = historyIndex - 1 < 0 ? inputHistory.length - 1 : historyIndex - 1
                    setHistoryIndex(newIndex)
                    setFormInput(inputHistory[newIndex])
                }
            }
        },
        [inputHistory, onChatSubmit, formInput, historyIndex, setFormInput]
    )

    const getBubbleClassName = (speaker: string): string => (speaker === 'human' ? 'human' : 'bot')

    useEffect(() => {
        if (transcriptContainerRef.current) {
            // Only scroll if the user didn't scroll up manually more than the scrolling threshold.
            // That is so that you can freely copy content or read up on older content while new
            // content is being produced.
            //
            // We allow some small threshold for "what is considered not scrolled up" so that
            // minimal scroll doesn't affect it (ie. if I'm not all the way scrolled down by like a
            // pixel or two, I probably still want it to scroll).
            //
            // Since this container uses flex-direction: column-reverse, the scrollTop starts in the
            // negatives and ends at 0.
            if (transcriptContainerRef.current.scrollTop >= -SCROLL_THRESHOLD) {
                transcriptContainerRef.current.scrollTo({ behavior: 'smooth', top: 0 })
            }
        }
    }, [transcript, transcriptContainerRef])

    return (
        <div className={classNames(className, styles.innerContainer)}>
            <div
                ref={transcriptContainerRef}
                className={classNames(styles.transcriptContainer, transcriptContainerClassName)}
            >
                {/* Show Tips view if no conversation has happened */}
                {transcript.length === 0 && !messageInProgress && (
                    <Tips recommendations={tipsRecommendations} after={afterTips} />
                )}
                {transcript.length > 0 && (
                    <div className={styles.bubbleContainer}>
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
                                            message.speaker === 'human'
                                                ? humanBubbleContentClassName
                                                : botBubbleContentClassName
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
                                                <div
                                                    className={classNames(
                                                        styles.bubbleLoaderDot,
                                                        bubbleLoaderDotClassName
                                                    )}
                                                />
                                                <div
                                                    className={classNames(
                                                        styles.bubbleLoaderDot,
                                                        bubbleLoaderDotClassName
                                                    )}
                                                />
                                                <div
                                                    className={classNames(
                                                        styles.bubbleLoaderDot,
                                                        bubbleLoaderDotClassName
                                                    )}
                                                />
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
                )}
            </div>

            <form className={classNames(styles.inputRow, inputRowClassName)}>
                <TextArea
                    className={classNames(styles.chatInput, chatInputClassName)}
                    rows={inputRows}
                    value={formInput}
                    autoFocus={true}
                    required={true}
                    onInput={({ target }) => {
                        const { value } = target as HTMLInputElement
                        inputHandler(value)
                    }}
                    onKeyDown={onChatKeyDown}
                />

                <SubmitButton className={styles.submitButton} onClick={onChatSubmit} disabled={!!messageInProgress} />
            </form>
        </div>
    )
}

import React, { useCallback, useEffect, useRef, useState } from 'react'

import classNames from 'classnames'
import useResizeObserver from 'use-resize-observer'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'
import { ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { CodeBlocks } from './chat/CodeBlocks'
import { ContextFiles, FileLinkProps } from './chat/ContextFiles'
import { Tips } from './Tips'

import styles from './Chat.module.css'

interface ChatProps extends ChatClassNames {
    messageInProgress: ChatMessage | null
    transcript: ChatMessage[]
    formInput: string
    setFormInput: (input: string) => void
    inputHistory: string[]
    setInputHistory: (history: string[]) => void
    onSubmit: (text: string) => void
    submitButtonComponent: React.FunctionComponent<ChatUISubmitButtonProps>
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
    tipsRecommendations?: JSX.Element[]
    afterTips?: JSX.Element
    className?: string
}

interface ChatClassNames {
    transcriptContainerClassName?: string
    bubbleContentClassName?: string
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
    submitButtonComponent: SubmitButton,
    fileLinkComponent,
    tipsRecommendations,
    afterTips,
    className,
    transcriptContainerClassName,
    bubbleContentClassName,
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
            if (event.key === 'Enter' && !event.shiftKey && formInput && formInput.trim()) {
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

    return (
        <div className={classNames(className, styles.innerContainer)}>
            <div className={classNames(styles.transcriptContainer, transcriptContainerClassName)}>
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
                                    styles[`${getBubbleClassName(message.speaker)}BubbleRow`]
                                )}
                            >
                                <div className={styles.bubble}>
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
                <AutoResizableTextArea
                    value={formInput}
                    onChange={setFormInput}
                    className={classNames(styles.chatInput, chatInputClassName)}
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

interface AutoResizableTextAreaProps {
    value: string
    onChange: (value: string) => void
    onInput?: React.FormEventHandler<HTMLTextAreaElement>
    onKeyDown?: React.KeyboardEventHandler<HTMLTextAreaElement>
    className?: string
}

export const AutoResizableTextArea: React.FC<AutoResizableTextAreaProps> = ({
    value,
    onChange,
    onInput,
    onKeyDown,
    className,
}) => {
    const textAreaRef = useRef<HTMLTextAreaElement>(null)
    const { width = 0 } = useResizeObserver({ ref: textAreaRef })

    const adjustTextAreaHeight = useCallback((): void => {
        if (textAreaRef.current) {
            textAreaRef.current.style.height = '0px'
            const scrollHeight = textAreaRef.current.scrollHeight
            textAreaRef.current.style.height = `${scrollHeight}px`

            // Hide scroll if the textArea isn't overflowing.
            textAreaRef.current.style.overflowY = scrollHeight < 200 ? 'hidden' : 'auto'
        }
    }, [])

    const handleChange = (event: React.ChangeEvent<HTMLTextAreaElement>): void => {
        onChange(event.target.value)
        adjustTextAreaHeight()
    }

    useEffect(() => {
        adjustTextAreaHeight()
    }, [adjustTextAreaHeight, value, width])

    return (
        <textarea
            ref={textAreaRef}
            className={className}
            value={value}
            onChange={handleChange}
            rows={1}
            autoFocus={true}
            required={true}
            onKeyDown={onKeyDown}
            onInput={onInput}
        />
    )
}

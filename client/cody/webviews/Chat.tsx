/* eslint-disable react/no-array-index-key */
/* eslint-disable jsx-a11y/no-noninteractive-element-interactions */
/* eslint-disable jsx-a11y/no-static-element-interactions */
/* eslint-disable jsx-a11y/click-events-have-key-events */
import React, { useCallback, useEffect, useRef, useState } from 'react'

import { VSCodeButton, VSCodeTextArea } from '@vscode/webview-ui-toolkit/react'

import { renderMarkdown } from '@sourcegraph/cody-shared/src/chat/markdown'

import { Tips } from './Tips'
import { SubmitSvg } from './utils/icons'
import { ChatMessage } from './utils/types'
import { vscodeAPI } from './utils/VSCodeApi'

import './Chat.css'

import { CodeBlocks } from './Components/CodeBlocks'

const SCROLL_THRESHOLD = 15

interface ChatboxProps {
    messageInProgress: ChatMessage | null
    transcript: ChatMessage[]
    formInput: string
    setFormInput: (input: string) => void
    inputHistory: string[]
    setInputHistory: (history: string[]) => void
}

export const Chat: React.FunctionComponent<React.PropsWithChildren<ChatboxProps>> = ({
    messageInProgress,
    transcript,
    formInput,
    setFormInput,
    inputHistory,
    setInputHistory,
}) => {
    const [inputRows, setInputRows] = useState(5)
    const [historyIndex, setHistoryIndex] = useState(inputHistory.length)
    const transcriptContainerRef = useRef<HTMLDivElement>(null)

    const inputHandler = useCallback(
        (inputValue: string) => {
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

    const escapeHTML = (html: string): string => {
        const span = document.createElement('span')
        span.textContent = html
        return span.innerHTML
    }

    const onChatSubmit = useCallback(() => {
        // Submit chat only when input is not empty
        if (formInput !== undefined) {
            vscodeAPI.postMessage({ command: 'submit', text: escapeHTML(formInput) })
            setHistoryIndex(inputHistory.length + 1)
            setInputHistory([...inputHistory, formInput])
            setInputRows(5)
            setFormInput('')
        }
    }, [formInput, inputHistory, setFormInput, setInputHistory])

    const onChatKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>): void => {
            // Submit input on Enter press (without shift)
            // trim the formInput to make sure input value is not empty
            if (event.key === 'Enter' && !event.shiftKey && formInput.trim()) {
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

    const bubbleClassName = (speaker: string): string => (speaker === 'human' ? 'human' : 'bot')

    useEffect(() => {
        if (transcriptContainerRef.current) {
            // Only scroll if the user didn't scroll up manually more than the scrolling threshold.
            // That is so that you can freely copy content or read up on older content while new
            // content is being produced.
            // We allow some small threshold for "what is considered not scrolled up" so that minimal
            // scroll doesn't affect it (ie. if I'm not all the way scrolled down by like a pixel or two,
            // I probably still want it to scroll).
            // Sice this container uses flex-direction: column-reverse, the scrollTop starts in the negatives
            // and ends at 0.
            if (transcriptContainerRef.current.scrollTop >= -SCROLL_THRESHOLD) {
                transcriptContainerRef.current.scrollTo({ behavior: 'smooth', top: 0 })
            }
        }
    }, [transcript, transcriptContainerRef])

    return (
        <div className="inner-container">
            <div ref={transcriptContainerRef} className={`${transcript.length >= 1 ? '' : 'non-'}transcript-container`}>
                {/* Show Tips view if no conversation has happened */}
                {transcript.length === 0 && !messageInProgress && <Tips />}
                {transcript.length > 0 && (
                    <div className="bubble-container">
                        {transcript.map((message, index) => (
                            <div
                                key={`message-${index}`}
                                className={`bubble-row ${bubbleClassName(message.speaker)}-bubble-row`}
                            >
                                <div className={`bubble ${bubbleClassName(message.speaker)}-bubble`}>
                                    <div
                                        className={`bubble-content ${bubbleClassName(message.speaker)}-bubble-content`}
                                    >
                                        {message.displayText && <CodeBlocks displayText={message.displayText} />}
                                        {message.contextFiles && message.contextFiles.length > 0 && (
                                            <ContextFiles contextFiles={message.contextFiles} />
                                        )}
                                    </div>
                                    <div className={`bubble-footer ${bubbleClassName(message.speaker)}-bubble-footer`}>
                                        <div className="bubble-footer-timestamp">{`${
                                            message.speaker === 'assistant' ? 'Cody' : 'Me'
                                        } · ${message.timestamp}`}</div>
                                    </div>
                                </div>
                            </div>
                        ))}

                        {messageInProgress && messageInProgress.speaker === 'assistant' && (
                            <div className="bubble-row bot-bubble-row">
                                <div className="bubble bot-bubble">
                                    <div className="bubble-content bot-bubble-content">
                                        {messageInProgress.displayText ? (
                                            <p
                                                dangerouslySetInnerHTML={{
                                                    __html: renderMarkdown(messageInProgress.displayText),
                                                }}
                                            />
                                        ) : (
                                            <div className="bubble-loader">
                                                <div className="bubble-loader-dot" />
                                                <div className="bubble-loader-dot" />
                                                <div className="bubble-loader-dot" />
                                            </div>
                                        )}
                                    </div>
                                    <div className="bubble-footer bot-bubble-footer">
                                        <span>Cody is typing...</span>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
            <form className="input-row">
                <VSCodeTextArea
                    className="chat-input"
                    rows={inputRows}
                    name="user-query"
                    value={formInput}
                    autofocus={true}
                    required={true}
                    onInput={({ target }) => {
                        const { value } = target as HTMLInputElement
                        inputHandler(value)
                    }}
                    onKeyDown={onChatKeyDown}
                />
                <VSCodeButton
                    className="submit-button"
                    appearance="icon"
                    type="button"
                    onClick={onChatSubmit}
                    disabled={!!messageInProgress}
                >
                    <SubmitSvg />
                </VSCodeButton>
            </form>
        </div>
    )
}

export const ContextFiles: React.FunctionComponent<{ contextFiles: string[] }> = ({ contextFiles }) => {
    const [isExpanded, setIsExpanded] = useState(false)

    if (contextFiles.length === 1) {
        return (
            <p>
                Cody read <code className="context-file">{contextFiles[0].split('/').pop()}</code> file to provide an
                answer.
            </p>
        )
    }

    if (isExpanded) {
        return (
            <p className="context-files-expanded">
                <span className="context-files-toggle-icon" onClick={() => setIsExpanded(false)}>
                    <i className="codicon codicon-triangle-down" slot="start" />
                </span>
                <div>
                    <div className="context-files-list-title" onClick={() => setIsExpanded(false)}>
                        Cody read the following files to provide an answer:
                    </div>
                    <ul className="context-files-list-container">
                        {contextFiles.map(file => (
                            <li key={file}>
                                <code className="context-file">{file}</code>
                            </li>
                        ))}
                    </ul>
                </div>
            </p>
        )
    }

    return (
        <p className="context-files-collapsed" onClick={() => setIsExpanded(true)}>
            <span className="context-files-toggle-icon">
                <i className="codicon codicon-triangle-right" slot="start" />
            </span>
            <div className="context-files-collapsed-text">
                <span>
                    Cody read <code className="context-file">{contextFiles[0].split('/').pop()}</code> and{' '}
                    {contextFiles.length - 1} other {contextFiles.length > 2 ? 'files' : 'file'} to provide an answer.
                </span>
            </div>
        </p>
    )
}

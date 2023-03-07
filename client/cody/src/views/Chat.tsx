import { useCallback, useState } from 'react'

import { VSCodeButton, VSCodeTextArea } from '@vscode/webview-ui-toolkit/react'

import { ChatMessage } from '../commands/ChatViewProvider'

import { Tips } from './About'
import { SubmitSvg } from './utils/icons'
import './App.css'
import { MessageFromWebview, vscodeAPI } from './utils/vscodeAPI'

interface ChatboxProps {
    setInConversation: (submitted: boolean) => void
    setUserInput: (input: string) => void
    userInput?: string
    transcript: ChatMessage | null
    setTranscript: (transcript: ChatMessage | null) => void
    transcripts: ChatMessage[]
    setTranscripts: (transcripts: ChatMessage[]) => void
    codyIsTyping: boolean
    setCodyIsTyping: (typing: boolean) => void
}

// Conversation with Cody
export const Chat: React.FunctionComponent<React.PropsWithChildren<ChatboxProps>> = ({
    setInConversation,
    setUserInput,
    userInput,
    transcript,
    transcripts,
    setTranscript,
    setTranscripts,
    codyIsTyping,
    setCodyIsTyping,
}) => {
    const [inputRows, setInputRows] = useState(5)
    const [formInput, setFormInput] = useState('')

    const inputHandler = useCallback(
        (inputValue: string) => {
            const rowsCount = inputValue.match(/\n/g)?.length
            if (rowsCount) {
                setInputRows(rowsCount < 5 ? 5 : rowsCount > 25 ? 25 : rowsCount)
            } else {
                setInputRows(5)
            }
            setFormInput(inputValue)
        },
        [setFormInput]
    )

    const onChatKeyDown = async (event: React.KeyboardEvent<HTMLDivElement>): void => {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault()
            event.stopPropagation()
            await onChatSubmit()
        }
    }

    const onChatSubmit = useCallback(async () => {
        setCodyIsTyping(true)
        setInputRows(5)
        setUserInput(formInput)
        setInConversation(true)
        const chatMsg: ChatMessage = {
            speaker: 'you',
            displayText: formInput,
            timestamp: getShortTimestamp(),
        }
        setTranscript(chatMsg)
        setTranscripts([...transcripts, chatMsg])

        vscodeAPI.postMessage({
            command: 'submit',
            text: formInput,
        } as MessageFromWebview)

        if (formInput === '/reset') {
            setUserInput('')
            setInConversation(false)
            setTranscript(null)
        }
        setFormInput('')
    }, [setCodyIsTyping, setUserInput, formInput, setInConversation, setTranscript, setTranscripts, transcripts])

    const bubbleClassName = (speaker: string): string => (speaker === 'you' ? 'human' : 'bot')

    return (
        <div className="inner-container">
            <div className={`${transcripts.length >= 1 ? '' : 'non-'}transcript-container`}>
                {!transcripts || (!transcripts[0] && <Tips />)}
                {transcripts?.[0] && (
                    <div className="container-getting-started">
                        {transcripts.map(msg => (
                            <div
                                key={msg.timestamp}
                                className={`bubble-row ${bubbleClassName(msg.speaker)}-bubble-row`}
                            >
                                <div className={`bubble ${bubbleClassName(msg.speaker)}-bubble`}>
                                    <div className={`bubble-content ${bubbleClassName(msg.speaker)}-bubble-content`}>
                                        {msg.contextFiles?.[0] && (
                                            <p data-contextfiles={JSON.stringify(msg.contextFiles)}>
                                                {msg.contextFiles}
                                            </p>
                                        )}
                                        <p dangerouslySetInnerHTML={{ __html: msg.displayText }} />
                                        {/* {feedback} */}
                                    </div>
                                    <div className={`bubble-footer ${bubbleClassName(msg.speaker)}-bubble-footer`}>
                                        <span>{`${msg.speaker === 'bot' ? 'Cody' : 'Me'} · ${msg.timestamp}`}</span>
                                    </div>
                                </div>
                            </div>
                        ))}

                        {codyIsTyping && (
                            <div className="bubble-row bot-bubble-row">
                                <div className="bubble bot-bubble">
                                    <div className="bubble-content bot-bubble-content">
                                        <div className="bubble-loader">
                                            <div className="bubble-loader-dot" />
                                            <div className="bubble-loader-dot" />
                                            <div className="bubble-loader-dot" />
                                        </div>
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
                    resize="vertical"
                    disabled={codyIsTyping}
                    required={true}
                    onInput={({ target }) => {
                        const { value } = target as HTMLInputElement
                        inputHandler(value)
                    }}
                    onKeyDown={onChatKeyDown}
                />
                <VSCodeButton className="chat-send-btn" appearance="icon" type="button" onClick={onChatSubmit}>
                    <SubmitSvg />
                </VSCodeButton>
            </form>
        </div>
    )
}

export function getShortTimestamp(): string {
    const date = new Date()
    return `${padTimePart(date.getHours())}:${padTimePart(date.getMinutes())}`
}

function padTimePart(timePart: number): string {
    return timePart < 10 ? `0${timePart}` : timePart.toString()
}

import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './UserHistory.css'
import './Chat.css'

import { useCallback, useState } from 'react'

import { ContextFiles } from './Chat'
import { ChatHistory, ChatMessage } from './utils/types'
import { vscodeAPI } from './utils/VSCodeApi'

interface HistoryProps {
    userHistory: ChatHistory | null
}

export const UserHistory: React.FunctionComponent<React.PropsWithChildren<HistoryProps>> = ({ userHistory }) => {
    const [chatHistory, setChatHistory] = useState('')

    const onRemoveHistoryClick = useCallback(() => {
        vscodeAPI.postMessage({ command: 'removeHistory' })
        setChatHistory('removed')
    }, [chatHistory])

    return (
        <div className="inner-container">
            <div className="non-transcript-container">
                <div className="bubble-container">
                    {userHistory &&
                        [...Object.entries(userHistory)].reverse().map(chat => (
                            <div key={chat[0]} className="history-item-container">
                                <VSCodeButton
                                    className="history-btn"
                                    type="button"
                                    onClick={() => setChatHistory(chatHistory === chat[0] ? '' : chat[0])}
                                >
                                    {chat[0]}
                                </VSCodeButton>
                                {chatHistory === chat[0] && (
                                    <div className="history-convo-container">
                                        {chat[1].map((message: ChatMessage, index: number) => (
                                            <div key={index} className="history-bubble-container bubble-content">
                                                {message.displayText && (
                                                    <p dangerouslySetInnerHTML={{ __html: message.displayText }} />
                                                )}
                                                {message.contextFiles && message.contextFiles.length > 0 && (
                                                    <ContextFiles contextFiles={message.contextFiles} />
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        ))}

                    <div className="history-item-container">
                        <VSCodeButton
                            className="history-btn history-remove-btn"
                            type="button"
                            onClick={onRemoveHistoryClick}
                            disabled={userHistory === null || chatHistory === 'removed'}
                        >
                            {userHistory === null || chatHistory === 'removed'
                                ? 'No chat history found'
                                : 'Remove all chat history'}
                        </VSCodeButton>
                    </div>
                </div>
            </div>
        </div>
    )
}

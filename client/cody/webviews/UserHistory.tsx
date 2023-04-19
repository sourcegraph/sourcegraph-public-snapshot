/* eslint-disable react/no-array-index-key */
import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import './UserHistory.css'

import { useCallback, useState } from 'react'

import classNames from 'classnames'

import { ChatHistory, ChatMessage } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { ContextFiles } from '@sourcegraph/cody-ui/src/chat/ContextFiles'

import { FileLink } from './FileLink'
import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './Chat.module.css'

interface HistoryProps {
    userHistory: ChatHistory | null
    setUserHistory: (history: ChatHistory | null) => void
    setInputHistory: (inputHistory: string[] | []) => void
    vscodeAPI: VSCodeWrapper
}

export const UserHistory: React.FunctionComponent<React.PropsWithChildren<HistoryProps>> = ({
    userHistory,
    setUserHistory,
    setInputHistory,
    vscodeAPI,
}) => {
    const [chatHistory, setChatHistory] = useState('')

    const onRemoveHistoryClick = useCallback(() => {
        if (userHistory) {
            vscodeAPI.postMessage({ command: 'removeHistory' })
            setChatHistory('removed')
            setUserHistory(null)
            setInputHistory([])
        }
    }, [setInputHistory, setUserHistory, userHistory, vscodeAPI])

    return (
        <div className={styles.innerContainer}>
            <div className={styles.nonTranscriptContainer}>
                <div className={styles.bubbleContainer}>
                    <div className="history-item-container">
                        <h3>Remove Chat & Input History</h3>
                        <VSCodeButton
                            className="history-btn history-remove-btn"
                            type="button"
                            appearance="secondary"
                            onClick={onRemoveHistoryClick}
                            disabled={userHistory === null || chatHistory === 'removed'}
                        >
                            {userHistory === null || chatHistory === 'removed'
                                ? 'Chat history is empty'
                                : 'Remove all local history'}
                        </VSCodeButton>
                        <h3>Local Chat History</h3>
                    </div>
                    {chatHistory !== 'removed' &&
                        userHistory &&
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
                                            <div
                                                key={index}
                                                className={classNames('history-bubble-container', styles.bubbleContent)}
                                            >
                                                {message.displayText && (
                                                    <p dangerouslySetInnerHTML={{ __html: message.displayText }} />
                                                )}
                                                {message.contextFiles && message.contextFiles.length > 0 && (
                                                    <ContextFiles
                                                        contextFiles={message.contextFiles}
                                                        fileLinkComponent={FileLink}
                                                    />
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        ))}
                </div>
            </div>
        </div>
    )
}

/* eslint-disable react/no-array-index-key */

import './UserHistory.css'

import { useCallback } from 'react'

import { ChatHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { View } from './NavBar'
import { VSCodeWrapper } from './utils/VSCodeApi'

import styles from './Chat.module.css'

interface HistoryProps {
    userHistory: ChatHistory | null
    setUserHistory: (history: ChatHistory | null) => void
    setInputHistory: (inputHistory: string[] | []) => void
    setView: (view: View | undefined) => void
    vscodeAPI: VSCodeWrapper
}

export const UserHistory: React.FunctionComponent<React.PropsWithChildren<HistoryProps>> = ({
    userHistory,
    setUserHistory,
    setInputHistory,
    setView,
    vscodeAPI,
}) => {
    const onRemoveHistoryClick = useCallback(() => {
        if (userHistory) {
            vscodeAPI.postMessage({ command: 'removeHistory' })
            setUserHistory(null)
            setInputHistory([])
        }
    }, [setInputHistory, setUserHistory, userHistory, vscodeAPI])

    function restoreMetadata(chatID: string): void {
        vscodeAPI.postMessage({ command: 'restoreHistory', chatID })
        setView('chat')
    }

    return (
        <div className={styles.innerContainer}>
            <div className={styles.nonTranscriptContainer}>
                <div className="history-item-container">
                    <h3>Remove Chat & Input History</h3>
                    <button
                        className="history-remove-btn"
                        type="button"
                        onClick={onRemoveHistoryClick}
                        disabled={userHistory === null}
                    >
                        {userHistory === null || Object.keys(userHistory).length === 0
                            ? 'Chat history is empty'
                            : 'Remove all local history'}
                    </button>
                    <h3>Local Chat History</h3>
                </div>
                {userHistory &&
                    [...Object.entries(userHistory)]
                        .sort(
                            (a, b) =>
                                +new Date(b[1].lastInteractionTimestamp) - +new Date(a[1].lastInteractionTimestamp)
                        )
                        .map(chat => {
                            const lastMessage = chat[1].interactions[chat[1].interactions.length - 1].assistantMessage
                            if (!lastMessage?.displayText) {
                                return null
                            }

                            return (
                                <div
                                    key={chat[0]}
                                    className="history-item"
                                    onClick={() => restoreMetadata(chat[0])}
                                    onKeyDown={event => {
                                        if (event.key === 'Enter') {
                                            restoreMetadata(chat[0])
                                        }
                                    }}
                                    role="button"
                                    tabIndex={0}
                                >
                                    <span className="history-item-date">
                                        <span>{new Date(chat[0]).toLocaleString()}</span>
                                    </span>
                                    <span className="history-item-last-message">
                                        {lastMessage.displayText.slice(0, 80)}
                                        {lastMessage.displayText.length > 80 ? '...' : ''}
                                    </span>
                                </div>
                            )
                        })}
            </div>
        </div>
    )
}

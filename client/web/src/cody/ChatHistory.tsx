import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/client'
import { Button, Text } from '@sourcegraph/wildcard'

import { safeTimestampToDate } from '../stores/codyChat'

import styles from './ChatHistory.module.scss'

interface ChatHistoryProps {
    transcriptHistory: TranscriptJSON[]
    loadTranscript: (id: string) => void
    closeHistory: () => void
    clearHistory: () => void
    showHeader?: boolean
    itemBodyClass?: string
    trucateMessageLenght?: number
}

export const ChatHistory: React.FunctionComponent<ChatHistoryProps> = ({
    transcriptHistory,
    loadTranscript,
    closeHistory,
    clearHistory,
    showHeader = true,
    itemBodyClass,
    trucateMessageLenght = 80,
}) => (
    <>
        {showHeader && (
            <Text className="p-2 pb-0" as="h3">
                Chat History
            </Text>
        )}
        {transcriptHistory.length === 0 && <Text className="p-2 pb-0 text-muted text-center">No chats yet</Text>}
        <ul className="p-0 d-flex flex-column">
            {transcriptHistory.reverse().map(({ id, interactions, lastInteractionTimestamp }) => {
                if (interactions.length === 0) {
                    return null
                }

                const lastInteraction = interactions[interactions.length - 1]
                const lastMessage = lastInteraction.assistantMessage || lastInteraction.humanMessage

                if (!lastMessage?.displayText) {
                    return null
                }

                /* eslint-disable jsx-a11y/no-noninteractive-element-to-interactive-role */
                return (
                    <li
                        role="button"
                        key={id}
                        className={styles.historyItem}
                        onClick={() => {
                            closeHistory()
                            loadTranscript(id)
                        }}
                        onKeyDown={() => {
                            closeHistory()
                            loadTranscript(id)
                        }}
                    >
                        <div className={`${styles.itemBody} ${itemBodyClass}`}>
                            <Text className="mb-1 text-muted" size="small">
                                <Timestamp date={safeTimestampToDate(lastInteractionTimestamp)} />
                            </Text>
                            <Text className="mb-0 truncate text-body">
                                {lastMessage.displayText.slice(0, trucateMessageLenght)}
                                {lastMessage.displayText.length > trucateMessageLenght ? '...' : ''}
                            </Text>
                        </div>
                    </li>
                )
                /* eslint-enable jsx-a11y/no-noninteractive-element-interactions */
            })}
        </ul>
        {showHeader && transcriptHistory.length > 0 && (
            <div className="text-center">
                <Button
                    variant="secondary"
                    outline={true}
                    onClick={() => {
                        closeHistory()
                        clearHistory()
                    }}
                >
                    Clear History
                </Button>
            </div>
        )}
    </>
)

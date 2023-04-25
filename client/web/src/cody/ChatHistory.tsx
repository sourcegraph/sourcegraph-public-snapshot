import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/client'
import { Button, Text } from '@sourcegraph/wildcard'

import styles from './ChatHistory.module.scss'

interface ChatHistoryProps {
    transcriptHistory: TranscriptJSON[]
    loadTranscript: (id: string) => void
    closeHistory: () => void
    clearHistory: () => void
}

export function ChatHistory({ transcriptHistory, loadTranscript, closeHistory, clearHistory }: ChatHistoryProps) {
    return (
        <>
            <Text className="p-2 pb-0" as="h3">
                Chat History
            </Text>
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

                    return (
                        <li
                            key={id}
                            className={styles.historyItem}
                            onClick={() => {
                                closeHistory()
                                loadTranscript(id)
                            }}
                        >
                            <div className={styles.itemBody}>
                                <Text className="mb-1 text-muted" size="small">
                                    <Timestamp date={new Date(lastInteractionTimestamp)} />
                                </Text>
                                <Text className="mb-0 truncate text-body">
                                    {lastMessage.displayText.slice(0, 80)}
                                    {lastMessage.displayText.length > 80 ? '...' : ''}
                                </Text>
                            </div>
                        </li>
                    )
                })}
            </ul>
            {transcriptHistory.length > 0 && (
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
}

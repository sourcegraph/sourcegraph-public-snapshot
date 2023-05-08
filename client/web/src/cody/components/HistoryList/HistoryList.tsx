import { useMemo } from 'react'

import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/transcript'
import { Text } from '@sourcegraph/wildcard'

import { safeTimestampToDate, useChatStoreState } from '../../stores/chat'

import styles from './HistoryList.module.scss'

interface HistoryListProps {
    trucateMessageLenght?: number
    onSelect?: (id: string) => void
}

export const HistoryList: React.FunctionComponent<HistoryListProps> = ({ trucateMessageLenght, onSelect }) => {
    const { transcriptHistory } = useChatStoreState()
    const transcripts = useMemo(
        () =>
            transcriptHistory.sort(
                (a, b) =>
                    -1 *
                    ((safeTimestampToDate(a.lastInteractionTimestamp) as any) -
                        (safeTimestampToDate(b.lastInteractionTimestamp) as any))
            ),
        [transcriptHistory]
    )

    return transcriptHistory.length === 0 ? (
        <Text className="p-2 pb-0 text-muted text-center">No chats yet</Text>
    ) : (
        <div className="p-0 d-flex flex-column">
            {transcripts.map(transcript => (
                <HistoryListItem
                    key={transcript.id}
                    transcript={transcript}
                    onSelect={onSelect}
                    truncateMessageLength={trucateMessageLenght}
                />
            ))}
        </div>
    )
}

const HistoryListItem: React.FunctionComponent<{
    transcript: TranscriptJSON
    truncateMessageLength?: number
    onSelect?: (id: string) => void
}> = ({ transcript: { id, interactions, lastInteractionTimestamp }, truncateMessageLength = 80, onSelect }) => {
    const { loadTranscriptFromHistory, transcriptId } = useChatStoreState()

    const lastMessage = useMemo(() => {
        let message = null

        for (let index = interactions.length - 1; index >= 0; index--) {
            const { assistantMessage, humanMessage } = interactions[index]

            if (assistantMessage?.text) {
                message = assistantMessage
            } else if (humanMessage?.text) {
                message = humanMessage
            }

            if (message) {
                break
            }
        }

        return message
    }, [interactions])

    if (!lastMessage?.text) {
        return null
    }

    return (
        <button
            key={id}
            type="button"
            className={classNames('text-left', styles.historyItem, {
                [styles.selected]: transcriptId === id,
            })}
            onClick={(): any => {
                onSelect?.(id)
                return loadTranscriptFromHistory(id)
            }}
            onKeyDown={(): any => {
                onSelect?.(id)
                return loadTranscriptFromHistory(id)
            }}
        >
            <Text className="mb-1 text-muted" size="small">
                <Timestamp date={safeTimestampToDate(lastInteractionTimestamp)} />
            </Text>
            <Text className="mb-0 truncate text-body">
                {lastMessage.text.slice(0, truncateMessageLength)}
                {lastMessage.text.length > truncateMessageLength ? '...' : ''}
            </Text>
        </button>
    )
}

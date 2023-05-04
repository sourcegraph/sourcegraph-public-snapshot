import { useMemo } from 'react'

import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Text } from '@sourcegraph/wildcard'

import { TranscriptJSON } from '../../../../../cody-shared/out/src/chat/transcript'
import { safeTimestampToDate, useChatStoreState } from '../../stores/chat'

import styles from './HistoryList.module.scss'

interface HistoryListProps {
    trucateMessageLenght?: number
    onSelect?: (id: string) => void
}

export const HistoryList: React.FunctionComponent<HistoryListProps> = ({ trucateMessageLenght, onSelect }) => {
    const { transcriptHistory } = useChatStoreState()

    return transcriptHistory.length === 0 ? (
        <Text className="p-2 pb-0 text-muted text-center">No chats yet</Text>
    ) : (
        <ul className="p-0 d-flex flex-column">
            {transcriptHistory
                .sort(
                    (a, b) =>
                        -1 *
                        ((safeTimestampToDate(a.lastInteractionTimestamp) as any) -
                            (safeTimestampToDate(b.lastInteractionTimestamp) as any))
                )
                .map(timestamp => (
                    <HistoryListItem
                        key={timestamp.id}
                        transcript={timestamp}
                        onSelect={onSelect}
                        truncateMessageLength={trucateMessageLenght}
                    />
                ))}
        </ul>
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

    /* eslint-disable jsx-a11y/no-noninteractive-element-to-interactive-role */
    return (
        <li
            role="button"
            key={id}
            className={classNames(styles.historyItem, {
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
        </li>
    )
    /* eslint-enable jsx-a11y/no-noninteractive-element-interactions */
}

import { useMemo, useCallback } from 'react'

import { mdiDelete } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { TranscriptJSON } from '@sourcegraph/cody-shared/src/chat/transcript'
import { Text, Icon, Tooltip } from '@sourcegraph/wildcard'

import { safeTimestampToDate, useChatStoreState } from '../../stores/chat'

import styles from './HistoryList.module.scss'

interface HistoryListProps {
    truncateMessageLength?: number
    onSelect?: (id: string) => void
    itemClassName?: string
}

export const HistoryList: React.FunctionComponent<HistoryListProps> = ({
    truncateMessageLength,
    onSelect,
    itemClassName,
}) => {
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
                    className={itemClassName}
                    truncateMessageLength={truncateMessageLength}
                />
            ))}
        </div>
    )
}

const HistoryListItem: React.FunctionComponent<{
    transcript: TranscriptJSON
    truncateMessageLength?: number
    onSelect?: (id: string) => void
    className?: string
}> = ({
    transcript: { id, interactions, lastInteractionTimestamp },
    truncateMessageLength = 80,
    onSelect,
    className,
}) => {
    const { loadTranscriptFromHistory, transcriptId, deleteHistoryItem } = useChatStoreState()

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

    const deleteItem = useCallback(
        (event: React.SyntheticEvent) => {
            event.stopPropagation()
            deleteHistoryItem(id)
        },
        [deleteHistoryItem, id]
    )

    const text = lastMessage?.text || 'Write your first message.'

    return (
        <button
            key={id}
            type="button"
            className={classNames(
                'text-left',
                styles.historyItem,
                {
                    [styles.selected]: transcriptId === id,
                },
                className
            )}
            onClick={(): any => {
                onSelect?.(id)
                return loadTranscriptFromHistory(id)
            }}
            onKeyDown={(): any => {
                onSelect?.(id)
                return loadTranscriptFromHistory(id)
            }}
        >
            <div className="d-flex align-items-center mb-1 justify-content-between w-100">
                <Text className="mb-1 text-muted" size="small">
                    <Timestamp date={safeTimestampToDate(lastInteractionTimestamp)} />
                </Text>
                <Tooltip content="Delete">
                    <Icon
                        aria-label="Delete"
                        svgPath={mdiDelete}
                        onClick={deleteItem}
                        className={styles.deleteButton}
                    />
                </Tooltip>
            </div>
            <Text className="mb-0 truncate text-body">
                {text.slice(0, truncateMessageLength)}
                {text.length > truncateMessageLength ? '...' : ''}
            </Text>
        </button>
    )
}

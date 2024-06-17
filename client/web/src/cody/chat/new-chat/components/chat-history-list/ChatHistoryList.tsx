import type { FC } from 'react'

import { mdiDelete, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { type ChatExportResult, getChatTitle } from '@sourcegraph/cody-web'
import { Icon, Text, Tooltip, Button } from '@sourcegraph/wildcard'

import styles from './ChatHistoryList.module.scss'

interface ChatHistoryListProps {
    chats: ChatExportResult[]
    isSelectedChat: (chat: ChatExportResult) => boolean
    className?: string
    onChatSelect: (chat: ChatExportResult) => void
    onChatDelete: (chat: ChatExportResult) => void
    onChatCreate: () => void
}

export const ChatHistoryList: FC<ChatHistoryListProps> = props => {
    const { chats, isSelectedChat, className, onChatSelect, onChatDelete, onChatCreate } = props

    return (
        <ul className={classNames(styles.historyList, className)}>
            {chats.map(chat => (
                <ChatHistoryItem
                    key={chat.chatID}
                    chat={chat}
                    selected={isSelectedChat(chat)}
                    onSelect={() => onChatSelect(chat)}
                    onDelete={() => onChatDelete(chat)}
                />
            ))}
            <Button variant="primary" onClick={onChatCreate} className="text-left">
                Start new chat
                <Icon aria-label="Add chat" svgPath={mdiPlus} />
            </Button>
        </ul>
    )
}

interface ChatHistoryItemProps {
    chat: ChatExportResult
    selected: boolean
    onSelect: () => void
    onDelete: () => void
}

const ChatHistoryItem: FC<ChatHistoryItemProps> = props => {
    const { chat, selected, onSelect, onDelete } = props
    const title = chat.transcript.chatTitle ?? getChatTitle(chat)
    const lastInteractionTimestamp = chat.transcript.lastInteractionTimestamp

    const handleDelete = (event: MouseEvent): void => {
        event.stopPropagation()
        onDelete()
    }

    return (
        <li>
            <button
                type="button"
                onClick={onSelect}
                className={classNames(styles.historyItem, { [styles.selected]: selected })}
            >
                <div className="d-flex align-items-center mb-1 justify-content-between w-100">
                    <Text className="mb-1 text-muted" size="small">
                        <Timestamp date={safeTimestampToDate(lastInteractionTimestamp)} />
                    </Text>
                    <Tooltip content="Delete">
                        <Icon
                            aria-label="Delete chat"
                            svgPath={mdiDelete}
                            onClick={handleDelete}
                            className={styles.deleteButton}
                        />
                    </Tooltip>
                </div>
                <Text className="mb-0 truncate text-body">{title}</Text>
            </button>
        </li>
    )
}

function safeTimestampToDate(timestamp: string = ''): Date {
    if (isNaN(new Date(timestamp) as any)) {
        return new Date()
    }

    return new Date(timestamp)
}

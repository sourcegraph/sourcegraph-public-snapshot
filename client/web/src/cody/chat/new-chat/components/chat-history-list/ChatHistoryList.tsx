import type { FC } from 'react'
import { type MouseEvent, useMemo } from 'react'

import { mdiDelete, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import type { ChatExportResult } from 'cody-web-experimental'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Icon, Text, Tooltip, Button } from '@sourcegraph/wildcard'

import styles from './ChatHistoryList.module.scss'

interface ChatHistoryListProps {
    chats: ChatExportResult[]
    isSelectedChat: (chat: ChatExportResult) => boolean
    className?: string
    onChatSelect: (chat: ChatExportResult) => void
    onChatDelete: (chat: ChatExportResult) => void
    onChatCreate: (force?: boolean) => void
}

export const ChatHistoryList: FC<ChatHistoryListProps> = props => {
    const { chats, isSelectedChat, className, onChatSelect, onChatDelete, onChatCreate } = props

    const sortedChats = useMemo(() => {
        try {
            const sortedChats = [...chats].sort(
                (chatA, chatB) =>
                    +safeTimestampToDate(chatB.transcript.lastInteractionTimestamp) -
                    +safeTimestampToDate(chatA.transcript.lastInteractionTimestamp)
            )

            const mostRecentEmptyChat = sortedChats.find(chat => chat.transcript.interactions.length === 0)

            if (mostRecentEmptyChat) {
                return [mostRecentEmptyChat, ...sortedChats.filter(chat => chat.transcript.interactions.length > 0)]
            }

            return sortedChats
        } catch {
            return chats
        }
    }, [chats])

    return (
        <ul className={classNames(styles.historyList, className)}>
            {sortedChats.map(chat => (
                <ChatHistoryItem
                    key={chat.chatID}
                    chat={chat}
                    selected={isSelectedChat(chat)}
                    onSelect={() => onChatSelect(chat)}
                    onDelete={() => onChatDelete(chat)}
                />
            ))}
            <footer className={styles.footer}>
                <Button variant="primary" onClick={() => onChatCreate()} className="w-100">
                    Start new chat
                    <Icon aria-label="Add chat" svgPath={mdiPlus} />
                </Button>
            </footer>
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
                <Text className={styles.historyItemTitle}>{title}</Text>
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

export function getChatTitle(chat: ChatExportResult): string {
    if (chat.transcript.chatTitle) {
        return chat.transcript.chatTitle
    }

    if (chat.transcript.interactions.length > 0) {
        const firstQuestion = chat.transcript.interactions.find(interaction => interaction.humanMessage.text)

        return firstQuestion?.humanMessage.text ?? ''
    }

    return 'New Cody Chat'
}

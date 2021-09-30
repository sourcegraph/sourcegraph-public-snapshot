import { sortBy } from 'lodash'
import React, { useState, useCallback } from 'react'
import stringScore from 'string-score'

import { HighlightedMatches } from '../../../components/HighlightedMatches'
import { Keybinding } from '../../../keyboardShortcuts'

import { NavigableList } from './NavigableList'

const KEEP_RECENT_ACTIONS = 10
const RECENT_ACTIONS_STORAGE_KEY = 'commandList.recentActions'

export interface CommandItem {
    id: string
    title: string
    keybindings?: Keybinding[]
    onClick: () => void
}

function readRecentActions(): string[] | null {
    const value = localStorage.getItem(RECENT_ACTIONS_STORAGE_KEY)
    if (value === null) {
        return null
    }
    try {
        const recentActions: unknown = JSON.parse(value)
        if (Array.isArray(recentActions) && recentActions.every(a => typeof a === 'string')) {
            return recentActions as string[]
        }
        return null
    } catch (error) {
        console.error('Error reading recent actions:', error)
    }
    writeRecentActions(null)
    return null
}

function writeRecentActions(recentActions: string[] | null): void {
    try {
        if (recentActions === null) {
            localStorage.removeItem(RECENT_ACTIONS_STORAGE_KEY)
        } else {
            const value = JSON.stringify(recentActions)
            localStorage.setItem(RECENT_ACTIONS_STORAGE_KEY, value)
        }
    } catch (error) {
        console.error('Error writing recent actions:', error)
    }
}

function filterAndRankItems(items: CommandItem[], query: string, recentActions: string[] | null): CommandItem[] {
    if (!query) {
        if (recentActions === null) {
            return items
        }
        // Show recent actions first.
        return sortBy(
            items,
            (item: Pick<CommandItem, 'id'>): number | null => {
                const index = recentActions.indexOf(item.id)
                return index === -1 ? null : index
            },
            ({ id }) => id
        )
    }

    // Memoize labels and scores.
    const labels = new Array<string>(items.length)
    const scores = new Array<number>(items.length)
    const scoredItems = items
        .filter((item, index) => {
            let label = labels[index]
            if (label === undefined) {
                label = item.title
                labels[index] = label
            }

            if (scores[index] === undefined) {
                scores[index] = stringScore(label, query, 0)
            }
            return scores[index] > 0
        })
        .map((item, index) => {
            const recentIndex = recentActions?.indexOf(item.id)
            return {
                item,
                score: scores[index],
                recentIndex: recentIndex === -1 ? null : recentIndex,
            }
        })
    return sortBy(scoredItems, 'recentIndex', 'score', ({ item }) => item.id).map(({ item }) => item)
}

interface CommandResultProps {
    value: string
    actions: CommandItem[]
    onClick: () => void
}

export const CommandResult: React.FC<CommandResultProps> = ({ actions, value, onClick: onClickProps }) => {
    const [recentActions, setRecentActions] = useState(readRecentActions)
    const filteredActions = actions && filterAndRankItems(actions, value, recentActions)

    const handleClick = useCallback(
        (id: string, onClick: () => void) => {
            onClickProps()
            onClick()
            setRecentActions(recentActions => {
                const newRecentActions = [id, ...(recentActions ?? [])].slice(0, KEEP_RECENT_ACTIONS)
                writeRecentActions(newRecentActions)
                return newRecentActions
            })
        },
        [onClickProps]
    )

    return (
        <NavigableList items={filteredActions}>
            {({ title, id, keybindings, onClick }, { active }) => (
                <NavigableList.Item
                    active={active}
                    keybindings={keybindings ?? []}
                    onClick={() => handleClick(id, onClick)}
                >
                    <HighlightedMatches text={title} pattern={value} />
                    {/* TODO add role=link and data-href for 'open'/'openPanel' commands,
                        render OpenInNewIcon */}
                </NavigableList.Item>
            )}
        </NavigableList>
    )
}

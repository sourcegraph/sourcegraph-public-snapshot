import { sortBy } from 'lodash'
import React, { useState, useCallback } from 'react'
import stringScore from 'string-score'

import { ActionItemAction } from '../../../actions/ActionItem'

const KEEP_RECENT_ACTIONS = 10
const RECENT_ACTIONS_STORAGE_KEY = 'commandList.recentActions'

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

function filterAndRankItems(
    items: Pick<ActionItemAction, 'action' | 'active'>[],
    query: string,
    recentActions: string[] | null
): ActionItemAction[] {
    if (!query) {
        if (recentActions === null) {
            return items
        }
        // Show recent actions first.
        return sortBy(
            items,
            (item: Pick<ActionItemAction, 'action'>): number | null => {
                const index = recentActions.indexOf(item.action.id)
                return index === -1 ? null : index
            },
            ({ action }) => action.id
        )
    }

    // Memoize labels and scores.
    const labels = new Array<string>(items.length)
    const scores = new Array<number>(items.length)
    const scoredItems = items
        .filter((item, index) => {
            let label = labels[index]
            if (label === undefined) {
                label = item.action.actionItem?.label
                    ? item.action.actionItem?.label
                    : `${item.action.category ? `${item.action.category}: ` : ''}${
                          item.action.title || item.action.command || ''
                      }`
                labels[index] = label
            }

            if (scores[index] === undefined) {
                scores[index] = stringScore(label, query, 0)
            }
            return scores[index] > 0
        })
        .map((item, index) => {
            const recentIndex = recentActions?.indexOf(item.action.id)
            return {
                item,
                score: scores[index],
                recentIndex: recentIndex === -1 ? null : recentIndex,
            }
        })
    return sortBy(scoredItems, 'recentIndex', 'score', ({ item }) => item.action.id).map(({ item }) => item)
}

interface CommandListResultProps {
    value: string
    onRunAction: (action: ActionItemAction) => void
    actions: ActionItemAction[]
    children: (label: string, keybindings: string[]) => JSX.Element
}

export const CommandListResult: React.FC<CommandListResultProps> = ({ actions, value, onRunAction }) => {
    const [recentActions, setRecentActions] = useState(readRecentActions)
    const filteredActions = actions && filterAndRankItems(actions, value, recentActions)

    console.log({ actions, recentActions, filteredActions })

    const handleRunAction = useCallback(
        (action: ActionItemAction) => {
            onRunAction(action)
            setRecentActions(recentActions => {
                const newRecentActions = [action.action.id, ...(recentActions ?? [])].slice(0, KEEP_RECENT_ACTIONS)
                writeRecentActions(newRecentActions)
                return newRecentActions
            })
        },
        [onRunAction]
    )

    return (
        <>
            {filteredActions?.map(item => {
                label = [action.category, action.actionItem?.label || action.title || action.command]
                    .filter(Boolean)
                    .join(': ')

                return children()
            })}
        </>
    )
}

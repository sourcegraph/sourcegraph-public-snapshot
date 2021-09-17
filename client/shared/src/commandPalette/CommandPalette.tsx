import Dialog from '@reach/dialog'
import { Remote } from 'comlink'
import * as H from 'history'
import { sortBy } from 'lodash'
import { extension } from 'mime-types'
import React, { useState, useMemo } from 'react'
import { useToggle } from 'react-use'
import { from } from 'rxjs'
import { filter, switchMap } from 'rxjs/operators'
import stringScore from 'string-score'

import { ActionItem, ActionItemAction } from '../actions/ActionItem'
import { wrapRemoteObservable } from '../api/client/api/common'
import { FlatExtensionHostAPI } from '../api/contract'
import { haveInitialExtensionsLoaded } from '../api/features'
import { ContributableMenu } from '../api/protocol'
import { HighlightedMatches } from '../components/HighlightedMatches'
import { getContributedActionItems } from '../contributions/contributions'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'
import { memoizeObservable } from '../util/memoizeObservable'
import { useObservable } from '../util/useObservable'

type CommandPaletteMode = 'fuzzy' | 'command' | 'jump-to-line'

function getMode(value: string): CommandPaletteMode {
    if (value.startsWith('>')) {
        return 'command'
    }
    if (value.startsWith(':')) {
        return 'jump-to-line'
    }
    // Default
    return 'fuzzy'
}

// Memoize contributions to prevent flashing loading spinners on subsequent mounts
const getContributions = memoizeObservable(
    (extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extensionHostAPI).pipe(switchMap(extensionHost => wrapRemoteObservable(extensionHost.getContributions()))),
    () => 'getContributions' // only one instance
)

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

export interface CommandPaletteProps
    extends ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps {
    isOpen?: boolean
    onDismiss?: () => void

    location: H.Location
}

const builtInActions: Pick<ActionItemAction, 'action' | 'active' | 'keybinding'>[] = [
    {
        action: {
            id: 'SOURCEGRAPH.switchColorTheme',
            actionItem: {
                label: 'Switch color theme',
            },
            command: 'open',
            commandArguments: ['https://google.com'],
        },
        keybinding: 'ctrl + t',
        active: true,
    },
]

/**
 * EXPERIMENTAL: New command palette (RFC 467)
 */
export const CommandPalette: React.FC<CommandPaletteProps> = ({
    isOpen = false,
    onDismiss,
    extensionsController,
    platformContext,
    telemetryService,
    location,
}) => {
    const [value, setValue] = useState('')

    const mode = getMode(value)

    const extensionContributions = useObservable(
        useMemo(
            () =>
                haveInitialExtensionsLoaded(extensionsController.extHostAPI).pipe(
                    // Don't listen for contributions until all initial extensions have loaded (to prevent UI jitter)
                    filter(haveLoaded => haveLoaded),
                    switchMap(() => getContributions(extensionsController.extHostAPI))
                ),
            [extensionsController]
        )
    )

    // Built in action items

    const actionItems = useMemo(
        () => [
            ...(extensionContributions
                ? getContributedActionItems(extensionContributions, ContributableMenu.CommandPalette)
                : []),
            ...builtInActions,
        ],
        // TODO: combine and map all actionItems
        [extensionContributions]
    )

    // TODO: recent actions, pass to filterAndRankItems

    const filteredActionItems = actionItems && filterAndRankItems(actionItems, value, null)

    return (
        // TODO: render shortcuts here. isOpen state is global, can be changed by e.g. button, keybinding.
        // this is a singleton component that is always rendered.
        <Dialog className="modal-body p-4 rounded border" {...{ isOpen, onDismiss }}>
            <div>
                <h1>cmdpal</h1>
                <input
                    autoComplete="off"
                    spellCheck="false"
                    aria-autocomplete="list"
                    className="form-control py-1"
                    placeholder="Search files by name (append : to jump to a line or @ to go to a symbol or > to search for a command)"
                    value={value}
                    onChange={event => {
                        setValue(event.target.value)
                    }}
                    type="text"
                />
            </div>
            {filteredActionItems?.map(item => (
                <li key={item.action.id}>
                    <ActionItem
                        {...item}
                        title={
                            <>
                                <HighlightedMatches
                                    text={[
                                        item.action.category,
                                        item.action.actionItem?.label || item.action.title || item.action.command,
                                    ]
                                        .filter(Boolean)
                                        .join(': ')}
                                    pattern={value}
                                />
                                {item.keybinding && <kbd>{item.keybinding}</kbd>}
                            </>
                        }
                        // onDidExecute={() => {}}
                        extensionsController={extensionsController}
                        platformContext={platformContext}
                        telemetryService={telemetryService}
                        location={location}
                    />
                </li>
            ))}
        </Dialog>
    )
}

export function filterAndRankItems(
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
            return { item, score: scores[index], recentIndex: recentIndex === -1 ? null : recentIndex }
        })
    return sortBy(scoredItems, 'recentIndex', 'score', ({ item }) => item.action.id).map(({ item }) => item)
}

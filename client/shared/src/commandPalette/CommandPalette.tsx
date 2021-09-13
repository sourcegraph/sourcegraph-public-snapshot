import Dialog from '@reach/dialog'
import { Remote } from 'comlink'
import React, { useState, useMemo } from 'react'
import { useToggle } from 'react-use'
import { from } from 'rxjs'
import { filter, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '../api/client/api/common'
import { FlatExtensionHostAPI } from '../api/contract'
import { haveInitialExtensionsLoaded } from '../api/features'
import { ExtensionsControllerProps } from '../extensions/controller'
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

export interface CommandPaletteProps extends ExtensionsControllerProps<'extHostAPI'> {
    defaultIsOpen?: boolean
}

/**
 * EXPERIMENTAL: New command palette (RFC 467)
 */
export const CommandPalette: React.FC<CommandPaletteProps> = ({ defaultIsOpen = false, extensionsController }) => {
    const [isOpen, toggleIsOpen] = useToggle(defaultIsOpen)
    const [value, setValue] = useState('')

    const mode = getMode(value)

    // observe extension commands
    const extensionCommands = useObservable(
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

    return (
        <div>
            <button type="button" onClick={toggleIsOpen}>
                Command Pallette
            </button>
            {isOpen && (
                <Dialog className="modal-body p-4 rounded border" isOpen={isOpen} onDismiss={toggleIsOpen}>
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
                </Dialog>
            )}
        </div>
    )
}

import Dialog from '@reach/dialog'
import { Remote } from 'comlink'
import * as H from 'history'
import React, { useState, useMemo, useCallback, useEffect } from 'react'
import { from } from 'rxjs'
import { filter, switchMap } from 'rxjs/operators'

import { ActionItemAction } from '../../actions/ActionItem'
import { wrapRemoteObservable } from '../../api/client/api/common'
import { FlatExtensionHostAPI } from '../../api/contract'
import { haveInitialExtensionsLoaded } from '../../api/features'
import { ContributableMenu } from '../../api/protocol'
import { getContributedActionItems } from '../../contributions/contributions'
import { ExtensionsControllerProps } from '../../extensions/controller'
import { PlatformContextProps } from '../../platform/context'
import { TelemetryProps } from '../../telemetry/telemetryService'
import { memoizeObservable } from '../../util/memoizeObservable'
import { useObservable } from '../../util/useObservable'

import { CommandListResult } from './components/CommandListResult'
import { CommandsModesList } from './components/CommandsModesList'
import { FuzzyFinderResult } from './components/FuzzyFinderResult'
import { JumpToLineResult } from './components/JumpToLineResult'
import { JumpToSymbolResult } from './components/JumpToSymbolResult'
import { RecentSearchesResult } from './components/RecentSearchesResult'
import { ShortcutController } from './components/ShortcutController'
import { useCommandPaletteStore } from './store'

export enum CommandPaletteMode {
    Fuzzy = '$',
    Command = '>',
    JumpToLine = ':',
    JumpToSymbol = '@',
    RecentSearches = '#',
}

const getMode = (text: string): CommandPaletteMode | undefined =>
    Object.values(CommandPaletteMode).find(value => text.startsWith(value))

// Memoize contributions to prevent flashing loading spinners on subsequent mounts
const getContributions = memoizeObservable(
    (extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extensionHostAPI).pipe(switchMap(extensionHost => wrapRemoteObservable(extensionHost.getContributions()))),
    () => 'getContributions' // only one instance
)

const BUILT_IN_ACTIONS: Pick<ActionItemAction, 'action' | 'active' | 'keybinding'>[] = [
    {
        action: {
            id: 'SOURCEGRAPH.switchColorTheme',
            actionItem: {
                label: 'Switch color theme',
            },
            command: 'open',
            commandArguments: ['https://google.com'],
        },
        keybinding: {
            ordered: ['T'],
            // held: ["Control"],
        },
        active: true,
    },
]

interface CommandPaletteActionItemProps {
    actionItem: ActionItemAction
    onRunAction: (action: ActionItemAction) => void
}

const CommandPaletteActionItem: React.FC<CommandPaletteActionItemProps> = ({ actionItem, onRunAction }) => {
    const { action, keybinding } = actionItem

    const label = [action.category, action.actionItem?.label || action.title || action.command]
        .filter(Boolean)
        .join(': ')

    return (
        <button type="button" onClick={(): void => onRunAction(actionItem)}>
            {label}

            {keybinding && (
                <>
                    {[...keybinding.ordered, ...(keybinding.held || [])].map(key => (
                        <kbd key={key}>{key}</kbd>
                    ))}
                </>
            )}
        </button>
    )
}

function useCommandList(value: string, extensionsController: CommandPaletteProps['extensionsController']) {
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

    const actions = useMemo(
        () => [
            ...(extensionContributions
                ? getContributedActionItems(extensionContributions, ContributableMenu.CommandPalette)
                : []),
            ...BUILT_IN_ACTIONS,
        ],
        // TODO: combine and map all actionItems
        [extensionContributions]
    )

    const onRunAction = useCallback((action: ActionItemAction) => {
        console.log('running action', action)
    }, [])

    const actionsWithShortcut = useMemo((): ActionItemAction[] => actions.filter(({ keybinding }) => !!keybinding), [
        actions,
    ])

    return { actions, actionsWithShortcut, onRunAction }
}

export interface CommandPaletteProps
    extends ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps {
    initialIsOpen?: boolean
    onDismiss?: () => void
    location: H.Location
}

/**
 * EXPERIMENTAL: New command palette (RFC 467)
 */
export const CommandPalette: React.FC<CommandPaletteProps> = ({
    initialIsOpen = false,
    // TODO: add ability to set default/initial mode
    onDismiss,
    extensionsController,
    platformContext,
    telemetryService,
    location,
}) => {
    const [text, setText] = useState('')
    const { actions, actionsWithShortcut, onRunAction } = useCommandList(text, extensionsController)

    const state = useCommandPaletteStore()
    const mode = getMode(text)

    useEffect(() => {
        if (initialIsOpen) {
            state.toggleIsOpen()
        }
        // Initial state for storybook
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const onClose = useCallback(() => {
        state.toggleIsOpen()
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const value = mode ? text.slice(1) : text

    return (
        // this is a singleton component that is always rendered.
        <>
            <ShortcutController actions={actionsWithShortcut} onMatch={onRunAction} />
            {state.isOpen && (
                <Dialog className="modal-body p-4 rounded border" isOpen={state.isOpen} onDismiss={onDismiss}>
                    <div>
                        <h1>cmdpal</h1>
                        <input
                            autoComplete="off"
                            spellCheck="false"
                            aria-autocomplete="list"
                            className="form-control py-1"
                            placeholder="Search files by name (append : to jump to a line or @ to go to a symbol or > to search for a command)"
                            value={text}
                            onChange={event => setText(event.target.value)}
                            type="text"
                        />
                    </div>
                    {!mode && <CommandsModesList />}
                    {mode === CommandPaletteMode.Command && (
                        <CommandListResult
                            actions={actions}
                            value={value}
                            onRunAction={action => {
                                onRunAction(action)
                                onClose()
                            }}
                        />
                    )}
                    {mode === CommandPaletteMode.RecentSearches && (
                        <RecentSearchesResult value={value} onClick={onClose} />
                    )}
                    {/* TODO: Only when repo open */}
                    {mode === CommandPaletteMode.Fuzzy && <FuzzyFinderResult value={value} onClick={onClose} />}
                    {/* TODO: Only when code editor open (possibly only when single open TODO) */}
                    {mode === CommandPaletteMode.JumpToLine && <JumpToLineResult value={value} onClick={onClose} />}
                    {mode === CommandPaletteMode.JumpToSymbol && <JumpToSymbolResult value={value} onClick={onClose} />}
                </Dialog>
            )}
        </>
    )
}

// import Dialog, { DialogContent, DialogOverlay } from '@reach/dialog'
import { Remote } from 'comlink'
import * as H from 'history'
import React, { useMemo, useCallback, useEffect } from 'react'
import { from, Observable } from 'rxjs'
import { filter, map, switchMap } from 'rxjs/operators'

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

import styles from './CommandPalette.module.scss'
import { CommandResult } from './components/CommandResult'
import { CommandsModesList } from './components/CommandsModesList'
import { FuzzyFinderResult } from './components/FuzzyFinderResult'
import { JumpToLineResult } from './components/JumpToLineResult'
import { JumpToSymbolResult } from './components/JumpToSymbolResult'
import { Modal } from './components/Modal'
import { RecentSearchesResult } from './components/RecentSearchesResult'
import { ShortcutController } from './components/ShortcutController'
import {
    COMMAND_PALETTE_SHORTCUTS,
    CommandPaletteMode,
    BUILT_IN_ACTIONS,
    KeyboardShortcutWithCallback,
} from './constants'
import { useCommandPaletteStore } from './store'

const getMode = (text: string): CommandPaletteMode | undefined =>
    Object.values(CommandPaletteMode).find(value => text.startsWith(value))

// Memoize contributions to prevent flashing loading spinners on subsequent mounts
const getContributions = memoizeObservable(
    (extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extensionHostAPI).pipe(switchMap(extensionHost => wrapRemoteObservable(extensionHost.getContributions()))),
    () => 'getContributions' // only one instance
)

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

    const shortcuts = useMemo((): KeyboardShortcutWithCallback[] => {
        const actionsWithShortcuts: KeyboardShortcutWithCallback[] = actions
            .filter(({ keybinding }) => !!keybinding)
            .map(action => ({
                keybindings: action.keybinding ? [action.keybinding] : [],
                onMatch: () => onRunAction(action),
                id: action.action.id,
                title: action.action.title ?? action.action.actionItem?.label ?? '',
            }))

        return [...COMMAND_PALETTE_SHORTCUTS, ...actionsWithShortcuts]
    }, [actions, onRunAction])

    return { actions, shortcuts, onRunAction }
}

export interface CommandPaletteProps
    extends ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'requestGraphQL'>,
        TelemetryProps {
    initialIsOpen?: boolean
    location: H.Location
    // TODO: different for web and bext. change name
    getAuthenticatedUserID: Observable<string | null>
}

/**
 * EXPERIMENTAL: New command palette (RFC 467)
 *
 * @description this is a singleton component that is always rendered.
 */
export const CommandPalette: React.FC<CommandPaletteProps> = ({
    initialIsOpen = false,
    // TODO: add ability to set default/initial mode
    extensionsController,
    platformContext,
    telemetryService,
    location,
    getAuthenticatedUserID,
}) => {
    const { isOpen, toggleIsOpen, value, setValue } = useCommandPaletteStore()
    const { actions, shortcuts, onRunAction } = useCommandList(value, extensionsController)

    const mode = getMode(value)

    useEffect(() => {
        if (initialIsOpen) {
            toggleIsOpen()
        }
    }, [toggleIsOpen, initialIsOpen])

    const handleClose = useCallback(() => {
        toggleIsOpen()
    }, [toggleIsOpen])

    const handleChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setValue(event.target.value)
        },
        [setValue]
    )

    const activeTextDocument = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getActiveTextDocument()))
                ),
            [extensionsController]
        )
    )

    const workspaceRoot = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getWorkspaceRoots())),
                    map(workspaceRoots => workspaceRoots[0])
                ),
            [extensionsController]
        )
    )

    const searchText = mode ? value.slice(1) : value

    return (
        <>
            <ShortcutController shortcuts={shortcuts} />
            {/* Can be rendered at the main app shell level */}
            <Modal.Host />
            {/* USE STYLES */}
            <Modal isOpen={isOpen} onDismiss={toggleIsOpen}>
                <Modal.Content>
                    <div>
                        <h1>cmdpal</h1>
                        <input
                            autoComplete="off"
                            spellCheck="false"
                            aria-autocomplete="list"
                            className="form-control py-1"
                            placeholder="Search files by name (append : to jump to a line or @ to go to a symbol or > to search for a command)"
                            value={value}
                            onChange={handleChange}
                            type="text"
                        />
                    </div>
                    {!mode && <CommandsModesList />}
                    {mode === CommandPaletteMode.Command && (
                        <CommandResult
                            actions={actions}
                            value={searchText}
                            onRunAction={action => {
                                onRunAction(action)
                                handleClose()
                            }}
                        />
                    )}
                    {mode === CommandPaletteMode.RecentSearches && (
                        <RecentSearchesResult
                            value={searchText}
                            onClick={handleClose}
                            getAuthenticatedUserID={getAuthenticatedUserID}
                            platformContext={platformContext}
                        />
                    )}
                    {/* TODO: Only when repo open */}
                    {mode === CommandPaletteMode.Fuzzy && (
                        <FuzzyFinderResult value={searchText} onClick={handleClose} workspaceRoot={workspaceRoot} />
                    )}
                    {/* TODO: Only when code editor open (possibly only when single open TODO) */}
                    {mode === CommandPaletteMode.JumpToLine && (
                        <JumpToLineResult
                            value={searchText}
                            onClick={handleClose}
                            textDocumentData={activeTextDocument}
                        />
                    )}
                    {mode === CommandPaletteMode.JumpToSymbol && (
                        <JumpToSymbolResult
                            value={searchText}
                            onClick={handleClose}
                            textDocumentData={activeTextDocument}
                            platformContext={platformContext}
                        />
                    )}
                </Modal.Content>
            </Modal>
        </>
    )
}

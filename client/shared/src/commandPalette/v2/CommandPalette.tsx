import { DialogContent, DialogOverlay } from '@reach/dialog'
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
import { CommandListResult } from './components/CommandListResult'
import { CommandsModesList } from './components/CommandsModesList'
import { FuzzyFinderResult } from './components/FuzzyFinderResult'
import { JumpToLineResult } from './components/JumpToLineResult'
import { JumpToSymbolResult } from './components/JumpToSymbolResult'
import { RecentSearchesResult } from './components/RecentSearchesResult'
import { ShortcutController } from './components/ShortcutController'
import { COMMAND_PALETTE_SHORTCUTS, CommandPaletteMode, BUILT_IN_ACTIONS } from './constants'
import { useCommandPaletteStore } from './store'

const getMode = (text: string): CommandPaletteMode | undefined =>
    Object.values(CommandPaletteMode).find(value => text.startsWith(value))

// Memoize contributions to prevent flashing loading spinners on subsequent mounts
const getContributions = memoizeObservable(
    (extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extensionHostAPI).pipe(switchMap(extensionHost => wrapRemoteObservable(extensionHost.getContributions()))),
    () => 'getContributions' // only one instance
)

interface CommandPaletteActionItemProps {
    actionItem: ActionItemAction
    onRunAction: (action: ActionItemAction) => void
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
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'requestGraphQL'>,
        TelemetryProps {
    initialIsOpen?: boolean
    location: H.Location
    // TODO: different for web and bext. change name
    getAuthenticatedUserID: Observable<string | null>
}

/**
 * EXPERIMENTAL: New command palette (RFC 467)
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
    const { actions, actionsWithShortcut, onRunAction } = useCommandList(value, extensionsController)

    const mode = getMode(value)

    useEffect(() => {
        if (initialIsOpen) {
            toggleIsOpen()
        }
    }, [toggleIsOpen, initialIsOpen])

    const onClose = useCallback(() => {
        toggleIsOpen()
    }, [toggleIsOpen])

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

    // TODO: merge other builtin shortcuts, extension-contributed  shortcuts
    const shortcuts = useMemo(() => [...COMMAND_PALETTE_SHORTCUTS], [])

    return (
        // this is a singleton component that is always rendered.
        <>
            <ShortcutController shortcuts={shortcuts} />
            {isOpen && (
                <DialogOverlay isOpen={isOpen} onDismiss={onClose} className={styles.dialogOverlay}>
                    <DialogContent className="modal-body p-4 rounded border shadow-lg">
                        <div>
                            <h1>cmdpal</h1>
                            <input
                                autoComplete="off"
                                spellCheck="false"
                                aria-autocomplete="list"
                                className="form-control py-1"
                                placeholder="Search files by name (append : to jump to a line or @ to go to a symbol or > to search for a command)"
                                value={value}
                                onChange={event => setValue(event.target.value)}
                                type="text"
                            />
                        </div>
                        {!mode && <CommandsModesList />}
                        {mode === CommandPaletteMode.Command && (
                            <CommandListResult
                                actions={actions}
                                value={searchText}
                                onRunAction={action => {
                                    onRunAction(action)
                                    onClose()
                                }}
                            />
                        )}
                        {mode === CommandPaletteMode.RecentSearches && (
                            <RecentSearchesResult
                                value={searchText}
                                onClick={onClose}
                                getAuthenticatedUserID={getAuthenticatedUserID}
                                platformContext={platformContext}
                            />
                        )}
                        {/* TODO: Only when repo open */}
                        {mode === CommandPaletteMode.Fuzzy && (
                            <FuzzyFinderResult value={searchText} onClick={onClose} workspaceRoot={workspaceRoot} />
                        )}
                        {/* TODO: Only when code editor open (possibly only when single open TODO) */}
                        {mode === CommandPaletteMode.JumpToLine && (
                            <JumpToLineResult
                                value={searchText}
                                onClick={onClose}
                                textDocumentData={activeTextDocument}
                            />
                        )}
                        {mode === CommandPaletteMode.JumpToSymbol && (
                            <JumpToSymbolResult
                                value={searchText}
                                onClick={onClose}
                                textDocumentData={activeTextDocument}
                            />
                        )}
                    </DialogContent>
                </DialogOverlay>
            )}
        </>
    )
}

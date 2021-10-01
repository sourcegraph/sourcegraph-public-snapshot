import { Remote } from 'comlink'
import * as H from 'history'
import React, { useMemo, useCallback, useEffect, useRef } from 'react'
import { Modal } from 'reactstrap'
import { from, Observable } from 'rxjs'
import { filter, map, switchMap } from 'rxjs/operators'

import { ActionItemAction, urlForClientCommandOpen } from '../../actions/ActionItem'
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
import { CommandPaletteModesResult } from './components/CommandPaletteModesResult'
import { CommandResult, CommandItem } from './components/CommandResult'
import { FuzzyFinderResult } from './components/FuzzyFinderResult'
import { InputField } from './components/InputField'
import { JumpToLineResult } from './components/JumpToLineResult'
import { JumpToSymbolResult } from './components/JumpToSymbolResult'
import { RecentSearchesResult } from './components/RecentSearchesResult'
import { ShortcutController, KeyboardShortcutWithCallback } from './components/ShortcutController'
import { COMMAND_PALETTE_COMMANDS, CommandPaletteMode } from './constants'
import { useCommandPaletteStore } from './store'

const getMode = (text: string): CommandPaletteMode | undefined =>
    Object.values(CommandPaletteMode).find(value => text.startsWith(value))

// Memoize contributions to prevent flashing loading spinners on subsequent mounts
const getContributions = memoizeObservable(
    (extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>) =>
        from(extensionHostAPI).pipe(switchMap(extensionHost => wrapRemoteObservable(extensionHost.getContributions()))),
    () => 'getContributions' // only one instance
)

// TEMPORARY MOCK FOR DEMO, REMOVE AFTER
const HACKATHON_DEMO_COMMANDS: CommandItem[] = [
    {
        id: 'githubCodeHost.newIssueWithSelection',
        title: 'GitHub: Open new issue in repository with selection',
        icon: 'https://upload.wikimedia.org/wikipedia/commons/9/91/Octicons-mark-github.svg',
        keybindings: [{ held: ['Alt'], ordered: ['I'] }],
        onClick: () => {
            const body = encodeURIComponent(
                'See selection:\n\nhttps://github.com/sourcegraph/sourcegraph/blob/c229c12a3a8eb82b3e62b1b8d7515b91641ca7f2/client/shared/src/api/extension/worker.ts#L9-L15'
            )

            window.open(`https://github.com/sourcegraph/sourcegraph/issues/new?body=${body}`, '_blank')
        },
    },
]

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
function useCommandList(extensionsController: CommandPaletteProps['extensionsController']) {
    const { extraCommands } = useCommandPaletteStore()

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

    const onRunAction = useCallback(
        (action: ActionItemAction['action']) => {
            if (!action.command) {
                // Unexpectedly arrived here; noop actions should not have event handlers that trigger
                // this.
                return
            }

            extensionsController
                .executeCommand({ command: action.command, args: action.commandArguments })
                .catch(error => console.error(error))
        },
        [extensionsController]
    )

    const extensionCommands: CommandItem[] = useMemo(() => {
        if (!extensionContributions) {
            return []
        }
        return getContributedActionItems(extensionContributions, ContributableMenu.CommandPalette).map(
            ({ action, keybinding }) => {
                const href = urlForClientCommandOpen(action, window.location)

                return {
                    id: action.id,
                    title: [action.category, action.title || action.command].filter(Boolean).join(': '),
                    keybindings: keybinding ? [keybinding] : [],
                    onClick: () => {
                        // Don't execute command since clicking on the link will essentially do the same thing.
                        if (!href) {
                            onRunAction(action)
                        }
                    },
                    icon: action.iconURL ?? action.actionItem?.iconURL,
                    href,
                }
            }
        )
    }, [extensionContributions, onRunAction])

    const shortcuts: KeyboardShortcutWithCallback[] = useMemo(
        () =>
            [...extensionCommands, ...COMMAND_PALETTE_COMMANDS, ...HACKATHON_DEMO_COMMANDS]
                .filter(({ keybindings }) => keybindings?.length)
                .map(({ id, keybindings = [], onClick, title }) => ({
                    keybindings,
                    onMatch: onClick,
                    id,
                    title,
                })),
        [extensionCommands]
    )

    const builtInCommands: CommandItem[] = useMemo(
        () => [...extraCommands, ...COMMAND_PALETTE_COMMANDS, ...HACKATHON_DEMO_COMMANDS],
        [extraCommands]
    )

    const actions = useMemo(() => [...extensionCommands, ...builtInCommands], [extensionCommands, builtInCommands])

    return { actions, shortcuts }
}

export interface CommandPaletteProps
    extends ExtensionsControllerProps,
        PlatformContextProps<
            'forceUpdateTooltip' | 'settings' | 'requestGraphQL' | 'clientApplication' | 'sourcegraphURL' | 'urlToFile'
        >,
        TelemetryProps {
    initialIsOpen?: boolean
    location: H.Location
    currentUserID: Observable<string | null>
}

/**
NOTE:

Mention existing (we learned)
- command palette
- fuzzy finder
- builtin actions (aka shortcuts)
- recent searches
- symbol mode

Why:
- Make those cool features ACCESSIBLE, NOTICABLE and better by grouping them in a single awesome cool UI
- Fast/quick navigation + make users more productive
- Creating a room for new built-in common commands/patterns
- Extending an extension API to set shortcut for commands

Future improvements:
- codehost integration + ability to customize and extend
- customizing shortcuts for each command

What we learned:
- value of prototyping
- a bit more about existing codebase (fuzzy finder, etc)
- zustand state management
- pair programming productivty boost
 */

/**
 * EXPERIMENTAL: New command palette (RFC 467)
 *
 * TODO: WRAP WITH ERROR BOUNDARY AT ALL CALL SITES
 *
 * @description this is a singleton component that is always rendered.
 */
export const CommandPalette: React.FC<CommandPaletteProps> = ({
    initialIsOpen = false,
    extensionsController,
    platformContext,
    currentUserID,
    telemetryService,
    location,
}) => {
    const { isOpen, toggleIsOpen, value, setValue } = useCommandPaletteStore()
    const { actions, shortcuts } = useCommandList(extensionsController)
    const inputReference = useRef<HTMLInputElement>(null)
    const mode = getMode(value)

    useEffect(() => {
        if (initialIsOpen) {
            toggleIsOpen()
        }
    }, [toggleIsOpen, initialIsOpen])

    const handleClose = useCallback(() => {
        toggleIsOpen()
    }, [toggleIsOpen])

    const handleInputFocus = useCallback(() => {
        inputReference.current?.focus()
    }, [])

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

    const placeholder = useMemo(
        () =>
            `Select mode by prefixing ${Object.values(CommandPaletteMode)
                .map(mode => `[${mode}]`)
                .join(' ')}`,
        []
    )
    const searchText = mode ? value.slice(1) : value

    return (
        <>
            <ShortcutController shortcuts={shortcuts} />
            {/* Can be rendered at the main app shell level */}

            {isOpen && (
                <Modal
                    isOpen={isOpen}
                    toggle={toggleIsOpen as () => void}
                    autoFocus={false}
                    backdropClassName={styles.modalBackdrop}
                    keyboard={true}
                    fade={false}
                    className={styles.modalDialog}
                    contentClassName={styles.modalContent}
                    returnFocusAfterClose={false}
                >
                    <InputField
                        ref={inputReference}
                        value={value}
                        onChange={setValue}
                        placeholder={placeholder}
                        isNative={platformContext.clientApplication === 'sourcegraph'}
                    />
                    {!mode && <CommandPaletteModesResult onSelect={handleInputFocus} />}
                    {mode === CommandPaletteMode.Command && (
                        <CommandResult actions={actions} value={searchText} onClick={handleClose} />
                    )}
                    {mode === CommandPaletteMode.RecentSearches && (
                        <RecentSearchesResult
                            value={searchText}
                            onClick={handleClose}
                            currentUserID={currentUserID}
                            platformContext={platformContext}
                        />
                    )}
                    {/* TODO: Only when repo open */}
                    {mode === CommandPaletteMode.Fuzzy && (
                        <FuzzyFinderResult
                            value={searchText}
                            onClick={handleClose}
                            workspaceRoot={workspaceRoot}
                            platformContext={platformContext}
                        />
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
                </Modal>
            )}
        </>
    )
}

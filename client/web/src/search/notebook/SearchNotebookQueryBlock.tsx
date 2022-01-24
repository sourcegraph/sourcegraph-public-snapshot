import classNames from 'classnames'
import { noop } from 'lodash'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useState, useCallback, useRef, useMemo, useEffect } from 'react'
import { useLocation } from 'react-router'
import { Observable, of, ReplaySubject } from 'rxjs'
import { filter, map, withLatestFrom } from 'rxjs/operators'

import { createHoverifier } from '@sourcegraph/codeintellify'
import { isDefined } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { StreamingSearchResultsList } from '@sourcegraph/search-ui'
import { useQueryDiagnostics } from '@sourcegraph/search/src/useQueryIntelligence'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { property } from '@sourcegraph/shared/src/util/types'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { getDocumentHighlights, getHover } from '../../backend/features'
import { WebHoverOverlay } from '../../components/WebHoverOverlay'
import { getLSPTextDocumentPositionParameters } from '../../repo/blob/Blob'
import { useExperimentalFeatures } from '../../stores'
import { SearchUserNeedsCodeHost } from '../../user/settings/codeHosts/OrgUserNeedsCodeHost'

import blockStyles from './SearchNotebookBlock.module.scss'
import { BlockMenuAction, SearchNotebookBlockMenu } from './SearchNotebookBlockMenu'
import styles from './SearchNotebookQueryBlock.module.scss'
import { useBlockSelection } from './useBlockSelection'
import { useBlockShortcuts } from './useBlockShortcuts'
import { useCommonBlockMenuActions } from './useCommonBlockMenuActions'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from './useMonacoBlockInput'

import { BlockProps, QueryBlock } from '.'

interface SearchNotebookQueryBlockProps
    extends BlockProps,
        QueryBlock,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        PlatformContextProps<'requestGraphQL' | 'urlToFile' | 'settings' | 'forceUpdateTooltip'>,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
    sourcegraphSearchLanguageId: string
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    authenticatedUser: AuthenticatedUser | null
}

export const SearchNotebookQueryBlock: React.FunctionComponent<SearchNotebookQueryBlockProps> = ({
    id,
    input,
    output,
    isLightTheme,
    telemetryService,
    settingsCascade,
    isSelected,
    isOtherBlockSelected,
    isMacPlatform,
    sourcegraphSearchLanguageId,
    fetchHighlightedFileLineRanges,
    onRunBlock,
    onSelectBlock,
    ...props
}) => {
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)

    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const blockElement = useRef<HTMLDivElement>(null)
    const searchResults = useObservable(output ?? of(undefined))
    const location = useLocation()

    const runBlock = useCallback(
        (id: string) => {
            if (!isSelected) {
                onSelectBlock(id)
            }
            onRunBlock(id)
        },
        [isSelected, onRunBlock, onSelectBlock]
    )

    const { isInputFocused } = useMonacoBlockInput({ editor, id, onRunBlock: runBlock, onSelectBlock, ...props })

    // setTimeout executes the editor focus in a separate run-loop which prevents adding a newline at the start of the input
    const onEnterBlock = useCallback(() => {
        setTimeout(() => editor?.focus(), 0)
    }, [editor])
    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        isSelected,
        isInputFocused,
        onSelectBlock,
        ...props,
    })
    const { onKeyDown } = useBlockShortcuts({ id, isMacPlatform, onEnterBlock, onRunBlock: runBlock, ...props })

    const modifierKeyLabel = isMacPlatform ? '⌘' : 'Ctrl'
    const mainMenuAction: BlockMenuAction = useMemo(() => {
        const isLoading = searchResults && searchResults.state === 'loading'
        return {
            type: 'button',
            label: isLoading ? 'Searching...' : 'Run search',
            isDisabled: isLoading ?? false,
            icon: <PlayCircleOutlineIcon className="icon-inline" />,
            onClick: runBlock,
            keyboardShortcutLabel: isSelected ? `${modifierKeyLabel} + ↵` : '',
        }
    }, [runBlock, isSelected, modifierKeyLabel, searchResults])

    const linkMenuActions: BlockMenuAction[] = useMemo(
        () => [
            {
                type: 'link',
                label: 'Open in new tab',
                icon: <OpenInNewIcon className="icon-inline" />,
                url: `/search?${buildSearchURLQuery(input, SearchPatternType.literal, false)}`,
            },
        ],
        [input]
    )

    const commonMenuActions = linkMenuActions.concat(
        useCommonBlockMenuActions({ modifierKeyLabel, isInputFocused, isMacPlatform, ...props })
    )

    useQueryDiagnostics(editor, { patternType: SearchPatternType.literal, interpretComments: true })

    // Element reference subjects passed to `hoverifier`
    const blockElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextBlockElement = useCallback((blockElement: HTMLElement | null) => blockElements.next(blockElement), [
        blockElements,
    ])

    const hoverOverlayElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextOverlayElement = useCallback(
        (overlayElement: HTMLElement | null) => hoverOverlayElements.next(overlayElement),
        [hoverOverlayElements]
    )

    // Subject that emits on every render. Source for `hoverOverlayRerenders`, used to
    // reposition hover overlay if needed when `SearchNotebook` rerenders
    const rerenders = useMemo(() => new ReplaySubject(1), [])
    useEffect(() => {
        rerenders.next()
    })

    // Create hoverifier.
    const hoverifier = useMemo(
        () =>
            createHoverifier<HoverContext, HoverMerged, ActionItemAction>({
                hoverOverlayElements,
                hoverOverlayRerenders: rerenders.pipe(
                    withLatestFrom(hoverOverlayElements, blockElements),
                    map(([, hoverOverlayElement, blockElement]) => ({
                        hoverOverlayElement,
                        relativeElement: blockElement,
                    })),
                    filter(property('relativeElement', isDefined)),
                    // Can't reposition HoverOverlay if it wasn't rendered
                    filter(property('hoverOverlayElement', isDefined))
                ),
                getHover: context =>
                    getHover(getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)), {
                        extensionsController: props.extensionsController,
                    }),
                getDocumentHighlights: context =>
                    getDocumentHighlights(
                        getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)),
                        { extensionsController: props.extensionsController }
                    ),
                getActions: context =>
                    getHoverActions(
                        { extensionsController: props.extensionsController, platformContext: props.platformContext },
                        context
                    ),
                tokenize: false,
            }),
        [
            // None of these dependencies are likely to change
            props.extensionsController,
            props.platformContext,
            hoverOverlayElements,
            blockElements,
            rerenders,
        ]
    )

    // Passed to HoverOverlay
    const hoverState = useObservable(hoverifier.hoverStateUpdates) || {}

    // Dispose hoverifier or change/unmount.
    useEffect(() => () => hoverifier.unsubscribe(), [hoverifier])

    return (
        <div
            className={classNames('block-wrapper', blockStyles.blockWrapper)}
            data-block-id={id}
            ref={nextBlockElement}
        >
            {/* Notebook blocks are a form of specialized UI for which there are no good accesibility settings (role, aria-*)
                or semantic elements that would accurately describe its functionality. To provide the necessary functionality we have
                to rely on plain div elements and custom click/focus/keyDown handlers. We still preserve the ability to navigate through blocks
                with the keyboard using the up and down arrows, and TAB. */}
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div
                className={classNames(
                    blockStyles.block,
                    styles.block,
                    isSelected && !isInputFocused && blockStyles.selected,
                    isSelected && isInputFocused && blockStyles.selectedNotFocused
                )}
                onClick={onSelect}
                onKeyDown={onKeyDown}
                onFocus={onSelect}
                // A tabIndex is necessary to make the block focusable.
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                aria-label="Notebook query block"
                ref={blockElement}
            >
                <div className="mb-1 text-muted">Search query</div>
                <div
                    className={classNames(
                        blockStyles.monacoWrapper,
                        isInputFocused && blockStyles.selected,
                        styles.queryInputMonacoWrapper
                    )}
                >
                    <MonacoEditor
                        language={sourcegraphSearchLanguageId}
                        value={input}
                        height="auto"
                        isLightTheme={isLightTheme}
                        editorWillMount={noop}
                        onEditorCreated={setEditor}
                        options={MONACO_BLOCK_INPUT_OPTIONS}
                        border={false}
                    />
                </div>

                {searchResults && searchResults.state === 'loading' && (
                    <div className={classNames('d-flex justify-content-center py-3', styles.results)}>
                        <LoadingSpinner />
                    </div>
                )}
                {searchResults && searchResults.state !== 'loading' && (
                    <div className={styles.results}>
                        <StreamingSearchResultsList
                            isSourcegraphDotCom={props.isSourcegraphDotCom}
                            searchContextsEnabled={props.searchContextsEnabled}
                            location={location}
                            allExpanded={false}
                            results={searchResults}
                            isLightTheme={isLightTheme}
                            fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                            telemetryService={telemetryService}
                            settingsCascade={settingsCascade}
                            authenticatedUser={props.authenticatedUser}
                            showSearchContext={showSearchContext}
                            assetsRoot={window.context?.assetsRoot || ''}
                            renderSearchUserNeedsCodeHost={user => <SearchUserNeedsCodeHost user={user} />}
                            platformContext={props.platformContext}
                            showCodeIntel={true}
                            extensionsController={props.extensionsController}
                            hoverifier={hoverifier}
                        />
                    </div>
                )}
            </div>
            {hoverState.hoverOverlayProps && (
                <WebHoverOverlay
                    {...props}
                    {...hoverState.hoverOverlayProps}
                    hoveredTokenElement={hoverState.hoveredTokenElement}
                    hoverRef={nextOverlayElement}
                    extensionsController={props.extensionsController}
                    location={location}
                    telemetryService={telemetryService}
                    isLightTheme={isLightTheme}
                />
            )}
            {(isSelected || !isOtherBlockSelected) && (
                <SearchNotebookBlockMenu
                    id={id}
                    mainAction={mainMenuAction}
                    actions={isSelected ? commonMenuActions : linkMenuActions}
                />
            )}
        </div>
    )
}

import React, { useState, useCallback, useMemo, useEffect } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import { Observable, of } from 'rxjs'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { SearchContextProps } from '@sourcegraph/search'
import { StreamingSearchResultsList, useQueryDiagnostics, FetchFileParameters } from '@sourcegraph/search-ui'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, useObservable, Icon } from '@sourcegraph/wildcard'

import { BlockProps, QueryBlock } from '../..'
import { AuthenticatedUser } from '../../../auth'
import { useExperimentalFeatures } from '../../../stores'
import { SearchUserNeedsCodeHost } from '../../../user/settings/codeHosts/OrgUserNeedsCodeHost'
import { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { focusLastPositionInMonacoEditor, useFocusMonacoEditorOnMount } from '../useFocusMonacoEditorOnMount'
import { useModifierKeyLabel } from '../useModifierKeyLabel'
import { MONACO_BLOCK_INPUT_OPTIONS, useMonacoBlockInput } from '../useMonacoBlockInput'

import blockStyles from '../NotebookBlock.module.scss'
import styles from './NotebookQueryBlock.module.scss'

interface NotebookQueryBlockProps
    extends BlockProps<QueryBlock>,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        ThemeProps,
        SettingsCascadeProps,
        TelemetryProps,
        PlatformContextProps<'requestGraphQL' | 'urlToFile' | 'settings' | 'forceUpdateTooltip'>,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
    isSourcegraphDotCom: boolean
    sourcegraphSearchLanguageId: string
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    authenticatedUser: AuthenticatedUser | null
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
}

export const NotebookQueryBlock: React.FunctionComponent<React.PropsWithChildren<NotebookQueryBlockProps>> = React.memo(
    ({
        id,
        input,
        output,
        isLightTheme,
        telemetryService,
        settingsCascade,
        isSelected,
        isOtherBlockSelected,
        sourcegraphSearchLanguageId,
        hoverifier,
        onBlockInputChange,
        fetchHighlightedFileLineRanges,
        onRunBlock,
        ...props
    }) => {
        const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
        const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
        const searchResults = useObservable(output ?? of(undefined))
        const [executedQuery, setExecutedQuery] = useState<string>(input.query)

        const onInputChange = useCallback(
            (query: string) => onBlockInputChange(id, { type: 'query', input: { query } }),
            [id, onBlockInputChange]
        )

        useEffect(() => {
            setExecutedQuery(input.query)
            // We intentionally want to track the input query state at the time
            // of search submission, not on input change.
            // eslint-disable-next-line react-hooks/exhaustive-deps
        }, [output])

        useMonacoBlockInput({
            editor,
            id,
            onRunBlock,
            onInputChange,
            ...props,
        })

        const modifierKeyLabel = useModifierKeyLabel()
        const mainMenuAction: BlockMenuAction = useMemo(() => {
            const isLoading = searchResults && searchResults.state === 'loading'
            return {
                type: 'button',
                label: isLoading ? 'Searching...' : 'Run search',
                isDisabled: isLoading ?? false,
                icon: <Icon role="img" aria-hidden={true} as={PlayCircleOutlineIcon} />,
                onClick: onRunBlock,
                keyboardShortcutLabel: isSelected ? `${modifierKeyLabel} + ↵` : '',
            }
        }, [onRunBlock, isSelected, modifierKeyLabel, searchResults])

        const linkMenuActions: BlockMenuAction[] = useMemo(
            () => [
                {
                    type: 'link',
                    label: 'Open in new tab',
                    icon: <Icon role="img" aria-hidden={true} as={OpenInNewIcon} />,
                    url: `/search?${buildSearchURLQuery(input.query, SearchPatternType.literal, false)}`,
                },
            ],
            [input]
        )

        const commonMenuActions = linkMenuActions.concat(useCommonBlockMenuActions({ id, ...props }))

        useQueryDiagnostics(editor, { patternType: SearchPatternType.literal, interpretComments: true })

        const focusInput = useCallback(() => focusLastPositionInMonacoEditor(editor), [editor])

        useFocusMonacoEditorOnMount({ editor, isEditing: input.initialFocusInput })

        return (
            <NotebookBlock
                className={styles.block}
                id={id}
                aria-label="Notebook query block"
                isSelected={isSelected}
                isOtherBlockSelected={isOtherBlockSelected}
                isInputVisible={true}
                focusInput={focusInput}
                mainAction={mainMenuAction}
                actions={isSelected ? commonMenuActions : linkMenuActions}
                {...props}
            >
                <div className="mb-1 text-muted">Search query</div>
                <div className={classNames(blockStyles.monacoWrapper, styles.queryInputMonacoWrapper)}>
                    <MonacoEditor
                        language={sourcegraphSearchLanguageId}
                        value={input.query}
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
                            extensionsController={props.extensionsController}
                            hoverifier={hoverifier}
                            openMatchesInNewTab={true}
                            executedQuery={executedQuery}
                        />
                    </div>
                )}
            </NotebookBlock>
        )
    }
)

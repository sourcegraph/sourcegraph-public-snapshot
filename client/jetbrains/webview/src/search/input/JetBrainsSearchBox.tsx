// This file is a fork from SearchBox.tsx and contains JetBrains specific UI changes
/* eslint-disable no-restricted-imports */

import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { type IEditor, LazyQueryInput } from '@sourcegraph/branded'
import { SearchContextDropdown } from '@sourcegraph/branded/src/search-ui/input/SearchContextDropdown'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { QueryState, SearchContextInputProps, SubmitSearchProps } from '@sourcegraph/shared/src/search'
import type { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Search } from '../jetbrains-icons/Search'

import { JetBrainsToggles, type JetBrainsTogglesProps } from './JetBrainsToggles'

import styles from './JetBrainsSearchBox.module.scss'

export interface JetBrainsSearchBoxProps
    extends Omit<JetBrainsTogglesProps, 'navbarSearchQuery' | 'submitSearch' | 'clearSearch'>,
        SearchContextInputProps,
        TelemetryProps,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean // significant for query suggestions
    showSearchContext: boolean
    showSearchContextManagement: boolean
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    submitSearchOnSearchContextChange?: SubmitSearchProps['submitSearch']
    submitSearchOnToggle?: SubmitSearchProps['submitSearch']
    onFocus?: () => void
    fetchStreamSuggestions?: typeof defaultFetchStreamSuggestions // Alternate implementation is used in the VS Code extension.
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    autoFocus?: boolean
    className?: string
    containerClassName?: string

    /** Whether comments are parsed and highlighted */
    interpretComments?: boolean

    /** Don't show search help button */
    hideHelpButton?: boolean

    /** Set in JSContext only available to the web app. */
    isExternalServicesUserModeAll?: boolean

    /** Called with the underlying editor instance on creation. */
    onEditorCreated?: (editor: IEditor) => void
}

export const JetBrainsSearchBox: React.FunctionComponent<React.PropsWithChildren<JetBrainsSearchBoxProps>> = props => {
    const { queryState, onEditorCreated: onEditorCreatedCallback, onChange } = props

    const [editor, setEditor] = useState<IEditor>()
    const focusEditor = useCallback(() => editor?.focus(), [editor])

    const onEditorCreated = useCallback(
        (editor: IEditor) => {
            setEditor(editor)
            onEditorCreatedCallback?.(editor)
        },
        [onEditorCreatedCallback]
    )

    const clearSearch = (): void => {
        onChange({ ...queryState, query: '' })
        focusEditor()
    }

    return (
        <div className={classNames(styles.searchBox, props.containerClassName)}>
            <div
                className={classNames(
                    styles.searchBoxBackgroundContainer,
                    props.className,
                    'flex-shrink-past-contents'
                )}
            >
                {props.searchContextsEnabled && props.showSearchContext && (
                    <>
                        <SearchContextDropdown
                            authenticatedUser={props.authenticatedUser}
                            // This is only used to render the CTA which we do not want on JetBrains
                            isSourcegraphDotCom={false}
                            searchContextsEnabled={props.searchContextsEnabled}
                            showSearchContextManagement={props.showSearchContextManagement}
                            setSelectedSearchContextSpec={props.setSelectedSearchContextSpec}
                            selectedSearchContextSpec={props.selectedSearchContextSpec}
                            fetchSearchContexts={props.fetchSearchContexts}
                            getUserSearchContextNamespaces={props.getUserSearchContextNamespaces}
                            telemetryService={props.telemetryService}
                            telemetryRecorder={props.telemetryRecorder}
                            platformContext={props.platformContext}
                            query={queryState.query}
                            submitSearch={props.submitSearchOnSearchContextChange}
                            className={classNames(styles.searchBoxContextDropdown, 'jb-search-context-dropdown')}
                            menuClassName={styles.searchBoxContextMenu}
                            onEscapeMenuClose={focusEditor}
                            ignoreDefaultContextDoesNotExistError={true}
                        />
                        <div className={styles.searchBoxSeparator} />
                    </>
                )}
                {/*
                    To fix Rule: "region" (All page content should be contained by landmarks)
                    Added role attribute to the following element to satisfy the rule.
                */}
                <div className={classNames(styles.searchBoxFocusContainer, 'flex-shrink-past-contents')} role="search">
                    <div className={styles.searchBoxFocusContainerIcon}>
                        <Search />
                    </div>
                    <LazyQueryInput
                        preventNewLine={true}
                        autoFocus={props.autoFocus}
                        caseSensitive={props.caseSensitive}
                        fetchStreamSuggestions={props.fetchStreamSuggestions}
                        isSourcegraphDotCom={props.isSourcegraphDotCom}
                        onChange={props.onChange}
                        onSubmit={props.onSubmit}
                        patternType={props.patternType}
                        queryState={props.queryState}
                        selectedSearchContextSpec={props.selectedSearchContextSpec}
                        className={styles.searchBoxInput}
                        onEditorCreated={onEditorCreated}
                        placeholder="Enter search query..."
                    />
                    <JetBrainsToggles
                        patternType={props.patternType}
                        setPatternType={props.setPatternType}
                        caseSensitive={props.caseSensitive}
                        setCaseSensitivity={props.setCaseSensitivity}
                        settingsCascade={props.settingsCascade}
                        submitSearch={props.submitSearchOnToggle}
                        navbarSearchQuery={queryState.query}
                        className={styles.searchBoxToggles}
                        structuralSearchDisabled={props.structuralSearchDisabled}
                        clearSearch={clearSearch}
                    />
                </div>
            </div>
        </div>
    )
}

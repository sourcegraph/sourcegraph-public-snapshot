import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'

import {
    SearchContextInputProps,
    QueryState,
    SubmitSearchProps,
    EditorHint,
    isSearchContextSpecAvailable,
} from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { IEditor, LazyMonacoQueryInput, LazyMonacoQueryInputProps } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { SearchContextDropdown } from './SearchContextDropdown'
import { SearchHelpDropdownButton } from './SearchHelpDropdownButton'
import { SearchHistoryDropdown } from './SearchHistoryDropdown'
import { Toggles, TogglesProps } from './toggles'

import styles from './SearchBox.module.scss'

export interface SearchBoxProps
    extends Omit<TogglesProps, 'navbarSearchQuery' | 'submitSearch'>,
        ThemeProps,
        SearchContextInputProps,
        TelemetryProps,
        PlatformContextProps<'requestGraphQL'>,
        Pick<
            LazyMonacoQueryInputProps,
            | 'editorComponent'
            | 'autoFocus'
            | 'onFocus'
            | 'onSubmit'
            | 'globbing'
            | 'interpretComments'
            | 'onChange'
            | 'onCompletionItemSelected'
            | 'applySuggestionsOnEnter'
            | 'defaultSuggestionsShowWhenEmpty'
            | 'showSuggestionsOnFocus'
        > {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean // significant for query suggestions
    showSearchContext: boolean
    showSearchContextManagement: boolean
    queryState: QueryState
    submitSearchOnSearchContextChange?: SubmitSearchProps['submitSearch']
    submitSearchOnToggle?: SubmitSearchProps['submitSearch']
    fetchStreamSuggestions?: typeof defaultFetchStreamSuggestions // Alternate implementation is used in the VS Code extension.
    className?: string
    containerClassName?: string

    /** Don't show search help button */
    hideHelpButton?: boolean

    /** Set in JSContext only available to the web app. */
    isExternalServicesUserModeAll?: boolean

    /** Called with the underlying editor instance on creation. */
    onEditorCreated?: (editor: IEditor) => void

    /** Whether or not to show the search history button. Also disables the
     * search button. Does not affect history in the search input itself (via
     * arrow up/down)
     */
    showSearchHistory?: boolean

    recentSearches?: RecentSearch[]
}

export const SearchBox: React.FunctionComponent<React.PropsWithChildren<SearchBoxProps>> = props => {
    const {
        queryState,
        onEditorCreated: onEditorCreatedCallback,
        showSearchHistory,
        hideHelpButton,
        onChange,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        platformContext,
        recentSearches,
    } = props

    const [editor, setEditor] = useState<IEditor>()
    const focusEditor = useCallback(() => editor?.focus(), [editor])

    const onEditorCreated = useCallback(
        (editor: IEditor) => {
            setEditor(editor)
            onEditorCreatedCallback?.(editor)
        },
        [onEditorCreatedCallback]
    )

    const onSearchHistorySelect = useCallback(
        (search: RecentSearch) => {
            const searchContext = getGlobalSearchContextFilter(search.query)
            if (searchContext) {
                console.log('selected')
                isSearchContextSpecAvailable({ spec: searchContext.spec, platformContext }).subscribe(isAvailable => {
                    console.log('available')
                    const query = isAvailable ? omitFilter(search.query, searchContext?.filter) : search.query
                    if (isAvailable) {
                        setSelectedSearchContextSpec(searchContext.spec)
                    }
                    onChange({ query, hint: EditorHint.Focus })
                })
            } else {
                onChange({ query: search.query, hint: EditorHint.Focus })
            }
        },
        [onChange, selectedSearchContextSpec, platformContext]
    )

    const historyWithoutSelectedContext = useMemo(() => {
        if (!recentSearches || !selectedSearchContextSpec) {
            return undefined
        }

        return recentSearches.map(search => {
            const searchContext = getGlobalSearchContextFilter(search.query)
            if (searchContext && searchContext.spec === selectedSearchContextSpec) {
                return {
                    ...search,
                    query: omitFilter(search.query, searchContext?.filter),
                }
            }
            return search
        })
    }, [recentSearches, selectedSearchContextSpec])

    return (
        <div
            className={classNames(
                styles.searchBox,
                props.containerClassName,
                props.hideHelpButton ? styles.searchBoxShadow : null
            )}
            data-testid="searchbox"
        >
            <div
                className={classNames(
                    styles.searchBoxBackgroundContainer,
                    props.showSearchHistory ? styles.searchBoxBackgroundContainerWithoutSearchButton : null,
                    props.className,
                    'flex-shrink-past-contents'
                )}
            >
                {showSearchHistory && (
                    <>
                        <SearchHistoryDropdown
                            recentSearches={props.recentSearches ?? []}
                            onSelect={onSearchHistorySelect}
                            onClose={focusEditor}
                        />
                        <div className={styles.searchBoxSeparator} />
                    </>
                )}
                {props.searchContextsEnabled && props.showSearchContext && (
                    <>
                        <SearchContextDropdown
                            {...props}
                            query={queryState.query}
                            submitSearch={props.submitSearchOnSearchContextChange}
                            className={styles.searchBoxContextDropdown}
                        />
                        <div className={styles.searchBoxSeparator} />
                    </>
                )}
                {/*
                    To fix Rule: "region" (All page content should be contained by landmarks)
                    Added role attribute to the following element to satisfy the rule.
                */}
                <div
                    className={classNames(
                        styles.searchBoxFocusContainer,
                        showSearchHistory ? styles.searchBoxFocusContainerWithoutSearchButton : null,
                        'flex-shrink-past-contents'
                    )}
                    role="search"
                >
                    <LazyMonacoQueryInput
                        className={styles.searchBoxInput}
                        onEditorCreated={onEditorCreated}
                        placeholder="Search for code and files"
                        preventNewLine={true}
                        autoFocus={props.autoFocus}
                        caseSensitive={props.caseSensitive}
                        editorComponent={props.editorComponent}
                        fetchStreamSuggestions={props.fetchStreamSuggestions}
                        globbing={props.globbing}
                        interpretComments={props.interpretComments}
                        isLightTheme={props.isLightTheme}
                        isSourcegraphDotCom={props.isSourcegraphDotCom}
                        onChange={props.onChange}
                        onCompletionItemSelected={props.onCompletionItemSelected}
                        onFocus={props.onFocus}
                        onSubmit={props.onSubmit}
                        patternType={props.patternType}
                        queryState={queryState}
                        selectedSearchContextSpec={props.selectedSearchContextSpec}
                        applySuggestionsOnEnter={props.applySuggestionsOnEnter}
                        defaultSuggestionsShowWhenEmpty={props.defaultSuggestionsShowWhenEmpty}
                        showSuggestionsOnFocus={props.showSuggestionsOnFocus}
                        searchHistory={historyWithoutSelectedContext}
                    />
                    <Toggles
                        patternType={props.patternType}
                        setPatternType={props.setPatternType}
                        caseSensitive={props.caseSensitive}
                        setCaseSensitivity={props.setCaseSensitivity}
                        settingsCascade={props.settingsCascade}
                        submitSearch={props.submitSearchOnToggle}
                        navbarSearchQuery={queryState.query}
                        className={styles.searchBoxToggles}
                        showCopyQueryButton={props.showCopyQueryButton}
                        structuralSearchDisabled={props.structuralSearchDisabled}
                        selectedSearchContextSpec={props.selectedSearchContextSpec}
                    />
                </div>
            </div>
            <SearchButton className={styles.searchBoxButton} />
            {!hideHelpButton && (
                <SearchHelpDropdownButton
                    isSourcegraphDotCom={props.isSourcegraphDotCom}
                    className={styles.helpButton}
                    telemetryService={props.telemetryService}
                />
            )}
        </div>
    )
}

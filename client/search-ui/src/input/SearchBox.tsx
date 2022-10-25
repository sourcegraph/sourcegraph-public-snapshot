import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { SearchContextInputProps, QueryState, SubmitSearchProps } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { IEditor, LazyMonacoQueryInput, LazyMonacoQueryInputProps } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { SearchContextDropdown } from './SearchContextDropdown'
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
            | 'suggestionSources'
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
}

export const SearchBox: React.FunctionComponent<React.PropsWithChildren<SearchBoxProps>> = props => {
    const { queryState, onEditorCreated: onEditorCreatedCallback } = props

    const [editor, setEditor] = useState<IEditor>()
    const focusEditor = useCallback(() => editor?.focus(), [editor])

    const onEditorCreated = useCallback(
        (editor: IEditor) => {
            setEditor(editor)
            onEditorCreatedCallback?.(editor)
        },
        [onEditorCreatedCallback]
    )

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
                    props.className,
                    'flex-shrink-past-contents'
                )}
            >
                {props.searchContextsEnabled && props.showSearchContext && (
                    <>
                        <SearchContextDropdown
                            {...props}
                            query={queryState.query}
                            submitSearch={props.submitSearchOnSearchContextChange}
                            className={styles.searchBoxContextDropdown}
                            onEscapeMenuClose={focusEditor}
                        />
                        <div className={styles.searchBoxSeparator} />
                    </>
                )}
                {/*
                    To fix Rule: "region" (All page content should be contained by landmarks)
                    Added role attribute to the following element to satisfy the rule.
                */}
                <div className={classNames(styles.searchBoxFocusContainer, 'flex-shrink-past-contents')} role="search">
                    <LazyMonacoQueryInput
                        className={styles.searchBoxInput}
                        onEditorCreated={onEditorCreated}
                        placeholder="Enter search query..."
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
                        suggestionSources={props.suggestionSources}
                        defaultSuggestionsShowWhenEmpty={props.defaultSuggestionsShowWhenEmpty}
                        showSuggestionsOnFocus={props.showSuggestionsOnFocus}
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
            <SearchButton
                hideHelpButton={props.hideHelpButton}
                className={styles.searchBoxButton}
                telemetryService={props.telemetryService}
                isSourcegraphDotCom={props.isSourcegraphDotCom}
            />
        </div>
    )
}

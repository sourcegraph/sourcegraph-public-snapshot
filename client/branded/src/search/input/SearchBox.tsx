import classNames from 'classnames'
import * as Monaco from 'monaco-editor'
import React, { useCallback, useState } from 'react'
import { Observable } from 'rxjs'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextInputProps } from '@sourcegraph/shared/src/search'
import { QueryState, SubmitSearchProps } from '@sourcegraph/shared/src/search/helpers'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import styles from './SearchBox.module.scss'
import { SearchButton } from './SearchButton'
import { SearchContextDropdown } from './SearchContextDropdown'
import { Toggles, TogglesProps } from './toggles/Toggles'

export interface SearchBoxProps
    extends Omit<TogglesProps, 'navbarSearchQuery' | 'submitSearch'>,
        ThemeProps,
        SearchContextInputProps,
        TelemetryProps,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean // significant for query suggestions
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    submitSearchOnSearchContextChange?: SubmitSearchProps['submitSearch']
    submitSearchOnToggle?: SubmitSearchProps['submitSearch']
    onFocus?: () => void
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut

    /** Whether globbing is enabled for filters. */
    globbing: boolean

    /** Whether comments are parsed and highlighted */
    interpretComments?: boolean

    /** Don't show search help button */
    hideHelpButton?: boolean

    onHandleFuzzyFinder?: React.Dispatch<React.SetStateAction<boolean>>

    /** Set in JSContext only available to the web app. */
    isExternalServicesUserModeAll?: boolean

    /** Fetch search suggestions for a given query. */
    fetchSuggestions?: (query: string) => Observable<SearchMatch[]>
}

export const SearchBox: React.FunctionComponent<SearchBoxProps> = props => {
    const { queryState, fetchSuggestions = fetchStreamSuggestions } = props

    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const focusEditor = useCallback(() => editor?.focus(), [editor])

    return (
        <div className={classNames(styles.searchBox, props.hideHelpButton ? styles.searchBoxShadow : null)}>
            <div className={classNames(styles.searchBoxBackgroundContainer, 'flex-shrink-past-contents')}>
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
                <div className={classNames(styles.searchBoxFocusContainer, 'flex-shrink-past-contents')}>
                    <LazyMonacoQueryInput
                        {...props}
                        onHandleFuzzyFinder={props.onHandleFuzzyFinder}
                        fetchSuggestions={fetchSuggestions}
                        className={styles.searchBoxInput}
                        onEditorCreated={setEditor}
                    />
                    <Toggles
                        {...props}
                        submitSearch={props.submitSearchOnToggle}
                        navbarSearchQuery={queryState.query}
                        className={styles.searchBoxToggles}
                    />
                </div>
            </div>
            <SearchButton
                hideHelpButton={props.hideHelpButton}
                className={styles.searchBoxButton}
                telemetryService={props.telemetryService}
                sourcegraphDotComMode={props.isSourcegraphDotCom}
            />
        </div>
    )
}

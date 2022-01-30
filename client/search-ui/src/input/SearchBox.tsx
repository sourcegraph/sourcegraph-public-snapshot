import classNames from 'classnames'
import * as Monaco from 'monaco-editor'
import React, { useCallback, useState } from 'react'

import { SearchContextInputProps, QueryState, SubmitSearchProps } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import styles from './SearchBox.module.scss'
import { SearchButton } from './SearchButton'
import { SearchContextDropdown } from './SearchContextDropdown'
import { Toggles, TogglesProps } from './toggles'

export interface SearchBoxProps
    extends Omit<TogglesProps, 'navbarSearchQuery' | 'submitSearch'>,
        ThemeProps,
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
    fetchStreamSuggestions?: typeof defaultFetchStreamSuggestions
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut
    className?: string
    containerClassName?: string

    /** Whether globbing is enabled for filters. */
    globbing: boolean

    /** Whether comments are parsed and highlighted */
    interpretComments?: boolean

    /** Don't show search help button */
    hideHelpButton?: boolean

    onHandleFuzzyFinder?: React.Dispatch<React.SetStateAction<boolean>>

    /** Set in JSContext only available to the web app. */
    isExternalServicesUserModeAll?: boolean
}

export const SearchBox: React.FunctionComponent<SearchBoxProps> = props => {
    const { queryState } = props

    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const focusEditor = useCallback(() => editor?.focus(), [editor])

    return (
        <div
            className={classNames(
                styles.searchBox,
                props.containerClassName,
                props.hideHelpButton ? styles.searchBoxShadow : null
            )}
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
                <div className={classNames(styles.searchBoxFocusContainer, 'flex-shrink-past-contents')}>
                    <LazyMonacoQueryInput
                        {...props}
                        onHandleFuzzyFinder={props.onHandleFuzzyFinder}
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
                isSourcegraphDotCom={props.isSourcegraphDotCom}
            />
        </div>
    )
}

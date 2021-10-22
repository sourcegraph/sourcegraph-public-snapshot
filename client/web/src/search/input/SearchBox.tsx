import classNames from 'classnames'
import React from 'react'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchContextInputProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { QueryState, SubmitSearchProps } from '../helpers'

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
        SettingsCascadeProps {
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
}

export const SearchBox: React.FunctionComponent<SearchBoxProps> = props => {
    const { queryState } = props

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
                        />
                        <div className={styles.searchBoxSeparator} />
                    </>
                )}
                <div className={classNames(styles.searchBoxFocusContainer, 'flex-shrink-past-contents')}>
                    <LazyMonacoQueryInput
                        {...props}
                        onHandleFuzzyFinder={props.onHandleFuzzyFinder}
                        className={styles.searchBoxInput}
                    />
                    <Toggles
                        {...props}
                        submitSearch={props.submitSearchOnToggle}
                        navbarSearchQuery={queryState.query}
                        className={styles.searchBoxToggles}
                    />
                </div>
            </div>
            <SearchButton hideHelpButton={props.hideHelpButton} className={styles.searchBoxButton} />
        </div>
    )
}

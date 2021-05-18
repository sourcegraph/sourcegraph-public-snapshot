import React from 'react'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CaseSensitivityProps, CopyQueryButtonProps, PatternTypeProps, SearchContextInputProps } from '..'
import { QueryState, submitSearch } from '../helpers'

import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import styles from './SearchBox.module.scss'
import { SearchContextDropdown } from './SearchContextDropdown'
import { Toggles, TogglesProps } from './toggles/Toggles'

export interface SearchBoxProps
    extends Omit<TogglesProps, 'navbarSearchQuery'>,
        ThemeProps,
        CaseSensitivityProps,
        PatternTypeProps,
        SearchContextInputProps,
        CopyQueryButtonProps {
    isSourcegraphDotCom: boolean // significant for query suggestions
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    onFocus?: () => void
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut
    submitSearchOnSearchContextChange?: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean

    // Whether to additionally highlight or provide hovers for tokens, e.g., regexp character sets.
    enableSmartQuery: boolean

    // Whether comments are parsed and highlighted
    interpretComments?: boolean
}

export const SearchBox: React.FunctionComponent<SearchBoxProps> = props => {
    const { queryState } = props

    return (
        <div className={styles.searchBox}>
            {props.showSearchContext && (
                <SearchContextDropdown query={queryState.query} submitSearch={submitSearch} {...props} />
            )}
            <div className={`${styles.searchBoxFocusContainer} flex-shrink-past-contents`}>
                <LazyMonacoQueryInput {...props} />
                <Toggles {...props} navbarSearchQuery={queryState.query} className={styles.searchBoxToggleContainer} />
            </div>
        </div>
    )
}

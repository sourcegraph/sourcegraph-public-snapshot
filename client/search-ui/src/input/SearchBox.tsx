import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { SearchContextInputProps, QueryState, SubmitSearchProps } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { LazyMonacoQueryInput, LazyMonacoQueryInputProps } from './LazyMonacoQueryInput'
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
        Pick<LazyMonacoQueryInputProps, 'editorComponent'> {
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

    /** Called with the underlying editor instance on creation. */
    onEditorCreated?: (editor: SearchEditor) => void
}

/**
 * Interface to the search box editor.
 * The underlying editor can be either Monaco or Codemirror.
 */
interface SearchEditor {
    focus: () => void
}

export const SearchBox: React.FunctionComponent<SearchBoxProps> = props => {
    const { queryState, onEditorCreated: onEditorCreatedCallback } = props

    const [editor, setEditor] = useState<SearchEditor>()
    const focusEditor = useCallback(() => editor?.focus(), [editor])

    const onEditorCreated = useCallback(
        (editor: SearchEditor) => {
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
                        onEditorCreated={onEditorCreated}
                        placeholder="Enter search query..."
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

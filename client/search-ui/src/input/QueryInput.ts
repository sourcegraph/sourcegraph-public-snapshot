import {
    CaseSensitivityProps,
    QueryState,
    SearchContextProps,
    SearchPatternTypeProps,
} from '@sourcegraph/shared/src/search'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { IEditor } from './LazyQueryInput'

/**
 * Props that the Monaco and CodeMirror implementation have in common.
 */
export interface QueryInputProps
    extends ThemeProps,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        SearchPatternTypeProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    isSourcegraphDotCom: boolean // Needed for query suggestions to give different options on dotcom; see SOURCEGRAPH_DOT_COM_REPO_COMPLETION
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit?: () => void
    onFocus?: () => void
    onBlur?: () => void
    onCompletionItemSelected?: () => void
    onEditorCreated?: (editor: IEditor) => void
    fetchStreamSuggestions?: typeof defaultFetchStreamSuggestions // Alternate implementation is used in the VS Code extension.
    autoFocus?: boolean
    // Whether globbing is enabled for filters.
    globbing: boolean

    // Whether comments are parsed and highlighted
    interpretComments?: boolean

    className?: string

    preventNewLine?: boolean

    /**
     * NOTE: This is currently only used for Insights code through
     * the MonacoField component: client/web/src/enterprise/insights/components/form/monaco-field/MonacoField.tsx
     *
     * Issue to improve this: https://github.com/sourcegraph/sourcegraph/issues/29438
     */
    placeholder?: string

    ariaLabel?: string
    ariaLabelledby?: string
}

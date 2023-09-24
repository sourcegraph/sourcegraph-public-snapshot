import type {
    CaseSensitivityProps,
    QueryState,
    SearchContextProps,
    SearchPatternTypeProps,
} from '@sourcegraph/shared/src/search'
import type { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

import type { IEditor } from './LazyQueryInput'

/**
 * Props that the Monaco and CodeMirror implementation have in common.
 */
export interface QueryInputProps
    extends Pick<CaseSensitivityProps, 'caseSensitive'>,
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

    // Whether comments are parsed and highlighted
    interpretComments?: boolean

    className?: string

    preventNewLine?: boolean

    placeholder?: string

    ariaLabel?: string
    ariaLabelledby?: string
    ariaInvalid?: string
    ariaBusy?: string
    tabIndex?: number
}

import type {
    CaseSensitivityProps,
    QueryState,
    SearchContextProps,
    SearchPatternTypeProps,
} from '@sourcegraph/shared/src/search'
import type { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

import type { IEditor as BaseEditor } from './LazyQueryInput'

/**
 * Props for the query input field.
 */
export interface QueryInputProps<E extends BaseEditor = BaseEditor>
    extends Pick<CaseSensitivityProps, 'caseSensitive'>,
        SearchPatternTypeProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    isSourcegraphDotCom: boolean // Needed for query suggestions to give different options on dotcom; see SOURCEGRAPH_DOT_COM_REPO_COMPLETION
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit?: () => void
    onFocus?: () => void
    onBlur?: () => void
    onEditorCreated?: (editor: E) => void
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

import { ReactElement, useCallback, useEffect, useMemo, useState } from 'react'

import { Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { Observable, of } from 'rxjs'
import { delay, startWith } from 'rxjs/operators'

import { CodeMirrorQueryInput, SyntaxHighlightedSearchQuery, singleLine, changeListener } from '@sourcegraph/branded'
import { pluralize } from '@sourcegraph/common'
import { createQueryExampleFromString, updateQueryWithFilterAndExample } from '@sourcegraph/shared/src/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { PathMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Label, useObservable } from '@sourcegraph/wildcard'

import { BlockProps } from '../..'
import { SearchPatternType } from '../../../graphql-operations'
import { blockKeymap, focusEditor } from '../../codemirror-utils'

import styles from './SearchTypeSuggestionsInput.module.scss'

interface SearchTypeSuggestionsInputProps<S extends SymbolMatch | PathMatch>
    extends ThemeProps,
        Pick<BlockProps, 'onRunBlock'> {
    id: string
    label: string
    queryPrefix: string
    queryInput: string
    onEditorCreated: (editor: EditorView) => void
    setQueryInput: (value: string) => void
    fetchSuggestions: (query: string) => Observable<S[]>
    countSuggestions: (suggestions: S[]) => number
    renderSuggestions: (suggestions: S[]) => ReactElement
    extension?: Extension
}

const LOADING = 'LOADING' as const
const QUERY_EXAMPLE = createQueryExampleFromString('{enter-regexp-pattern}')

export const SearchTypeSuggestionsInput = <S extends SymbolMatch | PathMatch>({
    id,
    label,
    queryPrefix,
    queryInput,
    setQueryInput,
    isLightTheme,
    fetchSuggestions,
    countSuggestions,
    renderSuggestions,
    onRunBlock,
    onEditorCreated,
    extension,
}: SearchTypeSuggestionsInputProps<S>): ReactElement => {
    const [editor, setEditor] = useState<EditorView | null>(null)

    const runBlock = useCallback(() => onRunBlock(id), [onRunBlock, id])
    const onEditorCreatedLocal = useCallback(
        (editor: EditorView) => {
            setEditor(editor)
            onEditorCreated(editor)
        },
        [onEditorCreated]
    )

    const addExampleFilter = useCallback(
        (filterType: FilterType) => {
            if (editor) {
                const { query, placeholderRange } = updateQueryWithFilterAndExample(
                    queryInput,
                    filterType,
                    QUERY_EXAMPLE,
                    { singular: false, negate: false, emptyValue: false }
                )
                editor.focus()
                editor.dispatch({
                    changes: { from: 0, to: editor.state.doc.length, insert: query },
                    selection: { anchor: placeholderRange.start, head: placeholderRange.end },
                    scrollIntoView: true,
                })
            }
        },
        [editor, queryInput]
    )

    const suggestions = useObservable(
        useMemo(
            () =>
                queryInput.length > 0
                    ? fetchSuggestions(queryInput).pipe(
                          // A small delay to prevent flickering loading message.
                          delay(300),
                          startWith(LOADING)
                      )
                    : of(undefined),
            [queryInput, fetchSuggestions]
        )
    )

    const suggestionsCount = useMemo(() => {
        if (!suggestions || suggestions === LOADING) {
            return undefined
        }
        return countSuggestions(suggestions)
    }, [suggestions, countSuggestions])

    // Focus input on component creation
    useEffect(() => {
        if (editor) {
            focusEditor(editor)
        }
    }, [editor])

    return (
        <div>
            <Label htmlFor={`${id}-search-type-query-input`}>{label}</Label>
            <div id={`${id}-search-type-query-input`} className={styles.queryInputWrapper}>
                <div className="d-flex">
                    <SyntaxHighlightedSearchQuery className={styles.searchTypeQueryPart} query={queryPrefix} />
                </div>
                <CodeMirrorQueryInput
                    value={queryInput}
                    isLightTheme={isLightTheme}
                    onEditorCreated={onEditorCreatedLocal}
                    patternType={SearchPatternType.standard}
                    interpretComments={true}
                    extensions={useMemo(
                        () => [blockKeymap({ runBlock }), changeListener(setQueryInput), singleLine, extension ?? []],
                        [setQueryInput, runBlock, extension]
                    )}
                />
            </div>
            <div className="mt-1">
                <Button
                    className="mr-1"
                    variant="secondary"
                    size="sm"
                    onClick={() => addExampleFilter(FilterType.repo)}
                >
                    Filter by repository
                </Button>
                <Button variant="secondary" size="sm" onClick={() => addExampleFilter(FilterType.file)}>
                    Filter by file path
                </Button>
            </div>
            <div className="mt-3 mb-1">
                {suggestionsCount !== undefined && (
                    <strong>
                        {suggestionsCount} {pluralize('result', suggestionsCount)} found
                    </strong>
                )}
                {suggestions === LOADING && <strong>Searching...</strong>}
            </div>
            {suggestions && suggestions !== LOADING && renderSuggestions(suggestions)}
        </div>
    )
}

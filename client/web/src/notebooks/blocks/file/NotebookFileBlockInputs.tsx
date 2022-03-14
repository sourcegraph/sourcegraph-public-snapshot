import classNames from 'classnames'
import { debounce } from 'lodash'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useCallback, useMemo, useState } from 'react'

import { isMacPlatform as isMacPlatformFn } from '@sourcegraph/common'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { PathMatch } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button } from '@sourcegraph/wildcard'

import { BlockProps, FileBlockInput } from '../..'
import { parseLineRange, serializeLineRange } from '../../serialize'
import { SearchTypeSuggestionsInput } from '../suggestions/SearchTypeSuggestionsInput'
import { fetchSuggestions } from '../suggestions/suggestions'

import styles from './NotebookFileBlockInputs.module.scss'

interface NotebookFileBlockInputsProps extends Pick<BlockProps, 'onSelectBlock' | 'onRunBlock'>, ThemeProps {
    id: string
    sourcegraphSearchLanguageId: string
    queryInput: string
    lineRange: IHighlightLineRange | null
    setQueryInput: (value: string) => void
    debouncedSetQueryInput: (value: string) => void
    setIsInputFocused(value: boolean): void
    onLineRangeChange: (lineRange: IHighlightLineRange | null) => void
    onFileSelected: (file: FileBlockInput) => void
}

function getFileSuggestionsQuery(queryInput: string): string {
    return `${queryInput} fork:yes type:path count:50`
}

export const NotebookFileBlockInputs: React.FunctionComponent<NotebookFileBlockInputsProps> = ({
    id,
    lineRange,
    setIsInputFocused,
    onSelectBlock,
    onFileSelected,
    onLineRangeChange,
    ...props
}) => {
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()
    const [lineRangeInput, setLineRangeInput] = useState(serializeLineRange(lineRange))
    const debouncedOnLineRangeChange = useMemo(() => debounce(onLineRangeChange, 300), [onLineRangeChange])

    const isLineRangeValid = useMemo(
        () => (lineRangeInput.trim() ? parseLineRange(lineRangeInput) !== null : undefined),
        [lineRangeInput]
    )

    const onLineRangeInputChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setLineRangeInput(event.target.value)
            debouncedOnLineRangeChange(parseLineRange(event.target.value))
        },
        [setLineRangeInput, debouncedOnLineRangeChange]
    )

    const onInputFocus = (event: React.FocusEvent<HTMLInputElement>): void => {
        onSelectBlock(id)
        setIsInputFocused(true)
        event.preventDefault()
        event.stopPropagation()
    }

    const onInputBlur = (event: React.FocusEvent<HTMLInputElement>): void => {
        // relatedTarget contains the element that will receive focus after the blur.
        const relatedTarget = event.relatedTarget && (event.relatedTarget as HTMLElement)
        // If relatedTarget is another input from the same block we
        // want to keep the input focused. Otherwise this will result in quickly flashing focus between elements.
        if (relatedTarget?.tagName.toLowerCase() !== 'input' && !relatedTarget?.closest('.monaco-editor')) {
            setIsInputFocused(false)
        }
        event.stopPropagation()
    }

    const fetchFileSuggestions = useCallback(
        (query: string) =>
            fetchSuggestions(
                getFileSuggestionsQuery(query),
                (suggestion): suggestion is PathMatch => suggestion.type === 'path',
                file => file
            ),
        []
    )

    const countSuggestions = useCallback((suggestions: PathMatch[]) => suggestions.length, [])

    const onFileSuggestionSelected = useCallback(
        (file: FileBlockInput) => {
            onFileSelected(file)
            setLineRangeInput(serializeLineRange(file.lineRange))
        },
        [onFileSelected, setLineRangeInput]
    )

    const renderSuggestions = useCallback(
        (suggestions: PathMatch[]) => (
            <FileSuggestions suggestions={suggestions} onFileSelected={onFileSuggestionSelected} />
        ),
        [onFileSuggestionSelected]
    )

    const isMacPlatform = useMemo(() => isMacPlatformFn(), [])

    return (
        <div className={styles.fileBlockInputs}>
            <div className="text-muted mb-2">
                <small>
                    <InfoCircleOutlineIcon className="icon-inline" /> To automatically select a file, copy a Sourcegraph
                    file URL, select the block, and paste the URL ({isMacPlatform ? '⌘' : 'Ctrl'} + v).
                </small>
            </div>
            <SearchTypeSuggestionsInput<PathMatch>
                id={id}
                editor={editor}
                setEditor={setEditor}
                setIsInputFocused={setIsInputFocused}
                onSelectBlock={onSelectBlock}
                label="Find a file using a Sourcegraph search query"
                queryPrefix="type:path"
                fetchSuggestions={fetchFileSuggestions}
                countSuggestions={countSuggestions}
                renderSuggestions={renderSuggestions}
                {...props}
            />
            <div className="mt-2">
                <label htmlFor={`${id}-line-range-input`}>Line range</label>
                <input
                    id={`${id}-line-range-input`}
                    type="text"
                    className={classNames('form-control', isLineRangeValid === false && 'is-invalid')}
                    value={lineRangeInput}
                    onChange={onLineRangeInputChange}
                    onBlur={onInputBlur}
                    onFocus={onInputFocus}
                    placeholder="Enter a single line (1), a line range (1-10), or leave empty to show the entire file."
                />
                {isLineRangeValid === false && (
                    <div className="text-danger mt-1">
                        Line range is invalid. Enter a single line (1), a line range (1-10), or leave empty to show the
                        entire file.
                    </div>
                )}
            </div>
        </div>
    )
}

const FileSuggestions: React.FunctionComponent<{
    suggestions: PathMatch[]
    onFileSelected: (symbol: FileBlockInput) => void
}> = ({ suggestions, onFileSelected }) => (
    <div className={styles.fileSuggestions}>
        {suggestions.map(suggestion => (
            <Button
                className={styles.fileButton}
                key={`${suggestion.repository}_${suggestion.path}`}
                onClick={() =>
                    onFileSelected({
                        repositoryName: suggestion.repository,
                        filePath: suggestion.path,
                        revision: suggestion.commit ?? '',
                        lineRange: null,
                    })
                }
                data-testid="file-suggestion-button"
            >
                <span className="mb-1">{suggestion.path}</span>
                <small className="text-muted">{suggestion.repository}</small>
            </Button>
        ))}
    </div>
)

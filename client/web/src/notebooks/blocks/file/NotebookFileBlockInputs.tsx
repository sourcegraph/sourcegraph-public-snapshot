import classNames from 'classnames'
import { debounce } from 'lodash'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import InfoCircleOutlineIcon from 'mdi-react/InfoCircleOutlineIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React, { useMemo } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { PathMatch, RepositoryMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { useObservable, Icon } from '@sourcegraph/wildcard'

import { BlockProps, FileBlockInput } from '../..'
import { parseLineRange } from '../../serialize'

import { NotebookFileBlockInput } from './NotebookFileBlockInput'
import styles from './NotebookFileBlockInputs.module.scss'
import { FileBlockInputValidationResult } from './useFileBlockInputValidation'

interface NotebookFileBlockInputsProps
    extends FileBlockInputValidationResult,
        Omit<FileBlockInput, 'lineRange'>,
        Pick<BlockProps, 'onSelectBlock'> {
    id: string
    lineRangeInput: string
    showRevisionInput: boolean
    showLineRangeInput: boolean
    isMacPlatform: boolean
    setIsInputFocused(value: boolean): void
    setFileInput: (input: Partial<FileBlockInput>) => void
    setLineRangeInput: (input: string) => void
}

const MAX_SUGGESTIONS = 15

function getRepositorySuggestionsQuery(repositoryName: string): string {
    return `repo:${repositoryName} type:repo count:${MAX_SUGGESTIONS} fork:yes`
}

function getFilePathSuggestionsQuery(repositoryName: string, revision: string, filePath: string): string {
    const repoFilter = repositoryName.trim() ? `repo:${repositoryName}` : ''
    const revisionFilter = revision.trim() ? `rev:${revision}` : ''
    return `${repoFilter} ${revisionFilter} ${filePath} type:path count:${MAX_SUGGESTIONS} fork:yes`
}

function fetchSuggestions<T extends RepositoryMatch | PathMatch>(
    query: string,
    filterSuggestionFn: (match: SearchMatch) => match is T,
    mapSuggestionFn: (match: T) => string
): Observable<string[]> {
    return fetchStreamSuggestions(query).pipe(
        map(suggestions => suggestions.filter(filterSuggestionFn).map(mapSuggestionFn))
    )
}

export const NotebookFileBlockInputs: React.FunctionComponent<NotebookFileBlockInputsProps> = ({
    id,
    repositoryName,
    filePath,
    revision,
    lineRangeInput,
    isRepositoryNameValid,
    isFilePathValid,
    isRevisionValid,
    isLineRangeValid,
    showRevisionInput,
    showLineRangeInput,
    isMacPlatform,
    setIsInputFocused,
    onSelectBlock,
    setFileInput,
    setLineRangeInput,
}) => {
    const onInputFocus = (event: React.FocusEvent<HTMLInputElement>): void => {
        onSelectBlock(id)
        setIsInputFocused(true)
        event.preventDefault()
        event.stopPropagation()
    }

    const debouncedSetFileInput = useMemo(() => debounce(setFileInput, 300), [setFileInput])

    const onInputBlur = (event: React.FocusEvent<HTMLInputElement>): void => {
        // relatedTarget contains the element that will receive focus after the blur.
        const relatedTarget = event.relatedTarget && (event.relatedTarget as HTMLElement)
        // If relatedTarget is another input from the same block or contained within the suggestions popover list, we
        // want to keep the input focused. Otherwise this will result in quickly flashing focus between elements.
        if (relatedTarget?.tagName.toLowerCase() !== 'input' && !relatedTarget?.closest('[data-reach-combobox-list]')) {
            setIsInputFocused(false)
        }
        event.stopPropagation()
    }

    const repoSuggestions = useObservable(
        useMemo(
            () =>
                fetchSuggestions(
                    getRepositorySuggestionsQuery(repositoryName),
                    (suggestion): suggestion is RepositoryMatch => suggestion.type === 'repo',
                    repo => repo.repository
                ),
            [repositoryName]
        )
    )

    const fileSuggestions = useObservable(
        useMemo(
            () =>
                fetchSuggestions(
                    getFilePathSuggestionsQuery(repositoryName, revision, filePath),
                    (suggestion): suggestion is PathMatch => suggestion.type === 'path',
                    repo => repo.path
                ),
            [repositoryName, revision, filePath]
        )
    )

    return (
        <div className={styles.fileBlockInputs}>
            <div className="text-muted mb-2">
                <small>
                    <Icon as={InfoCircleOutlineIcon} /> To automatically fill the inputs, copy a Sourcegraph file URL,
                    select the block, and paste the URL ({isMacPlatform ? '⌘' : 'Ctrl'} + v).
                </small>
            </div>
            <label htmlFor={`file-location-input-${id}`}>File location</label>
            <div id={`file-location-input-${id}`} className={styles.fileLocationInputWrapper}>
                <NotebookFileBlockInput
                    className="flex-1"
                    inputClassName={styles.repositoryNameInput}
                    value={repositoryName}
                    placeholder="Repository name"
                    onChange={repositoryName => debouncedSetFileInput({ repositoryName })}
                    onFocus={onInputFocus}
                    onBlur={onInputBlur}
                    suggestions={repoSuggestions}
                    suggestionsIcon={<SourceRepositoryIcon className="mr-1" size="1rem" />}
                    isValid={isRepositoryNameValid}
                    isMacPlatform={isMacPlatform}
                    dataTestId="file-block-repository-name-input"
                />
                <div className={styles.separator} />
                <NotebookFileBlockInput
                    className="flex-1"
                    inputClassName={styles.filePathInput}
                    value={filePath}
                    placeholder="Path"
                    onChange={filePath => debouncedSetFileInput({ filePath })}
                    onFocus={onInputFocus}
                    onBlur={onInputBlur}
                    suggestions={fileSuggestions}
                    suggestionsIcon={<FileDocumentIcon className="mr-1" size="1rem" />}
                    isValid={isFilePathValid}
                    isMacPlatform={isMacPlatform}
                    dataTestId="file-block-file-path-input"
                />
            </div>
            <div className={classNames('d-flex', (showRevisionInput || showLineRangeInput) && 'mt-3')}>
                {showRevisionInput && (
                    <div className="w-50 mr-2">
                        <label htmlFor={`file-revision-input-${id}`}>Revision</label>
                        <NotebookFileBlockInput
                            id={`file-revision-input-${id}`}
                            inputClassName={styles.revisionInput}
                            value={revision}
                            placeholder="feature/branch"
                            onChange={revision => debouncedSetFileInput({ revision })}
                            onFocus={onInputFocus}
                            onBlur={onInputBlur}
                            isValid={isRevisionValid}
                            isMacPlatform={isMacPlatform}
                            dataTestId="file-block-revision-input"
                        />
                    </div>
                )}
                {showLineRangeInput && (
                    <div className="w-50">
                        <label htmlFor={`file-line-range-input-${id}`}>Line range</label>
                        <NotebookFileBlockInput
                            id={`file-line-range-input-${id}`}
                            inputClassName={styles.lineRangeInput}
                            value={lineRangeInput}
                            placeholder="1-10"
                            onChange={lineRangeInput => {
                                setLineRangeInput(lineRangeInput)
                                const lineRange = parseLineRange(lineRangeInput)
                                if (lineRange !== null) {
                                    debouncedSetFileInput({ lineRange })
                                }
                            }}
                            onFocus={onInputFocus}
                            onBlur={onInputBlur}
                            isValid={isLineRangeValid}
                            isMacPlatform={isMacPlatform}
                            dataTestId="file-block-line-range-input"
                        />
                    </div>
                )}
            </div>
        </div>
    )
}

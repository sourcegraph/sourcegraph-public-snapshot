import { useMemo } from 'react'
import { Observable, of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'

import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { useObservable } from '@sourcegraph/wildcard'

import { FileBlockInput } from '..'
import { fetchRepository as _fetchRepository, resolveRevision as _resolveRevision } from '../../../repo/backend'
import { parseLineRange } from '../serialize'

function validateInput<T>(
    input: string,
    validationFn: (input: string) => Observable<T>
): Observable<boolean | undefined> {
    return input.trim()
        ? validationFn(input).pipe(
              map(() => true),
              catchError(() => of(false))
          )
        : of(undefined)
}

export interface FileBlockValidationFunctions {
    fetchRepository: typeof _fetchRepository
    resolveRevision: typeof _resolveRevision
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export interface FileBlockInputValidationResult {
    isRepositoryNameValid: boolean | undefined
    isFilePathValid: boolean | undefined
    isRevisionValid: boolean | undefined
    isLineRangeValid: boolean | undefined
}

export function useFileBlockInputValidation(
    input: Omit<FileBlockInput, 'lineRange'>,
    lineRangeInput: string,
    { fetchRepository, resolveRevision, fetchHighlightedFileLineRanges }: FileBlockValidationFunctions
): FileBlockInputValidationResult {
    const isRepositoryNameValid = useObservable(
        useMemo(() => validateInput(input.repositoryName, repoName => fetchRepository({ repoName })), [
            fetchRepository,
            input.repositoryName,
        ])
    )

    const isFilePathValid = useObservable(
        useMemo(
            () =>
                validateInput(input.filePath, filePath =>
                    fetchHighlightedFileLineRanges({
                        repoName: input.repositoryName,
                        commitID: input.revision,
                        filePath,
                        ranges: [],
                        disableTimeout: true,
                    })
                ),
            [input.repositoryName, input.filePath, input.revision, fetchHighlightedFileLineRanges]
        )
    )

    const isRevisionValid = useObservable(
        useMemo(
            () =>
                validateInput(input.revision, revision =>
                    resolveRevision({ repoName: input.repositoryName, revision })
                ),
            [input.repositoryName, input.revision, resolveRevision]
        )
    )

    const isLineRangeValid = useMemo(
        () => (lineRangeInput.trim() ? parseLineRange(lineRangeInput) !== null : undefined),
        [lineRangeInput]
    )

    return {
        isRepositoryNameValid,
        isFilePathValid,
        isRevisionValid,
        isLineRangeValid,
    }
}

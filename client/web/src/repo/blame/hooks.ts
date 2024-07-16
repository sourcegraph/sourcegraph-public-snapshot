import { useMemo, useCallback } from 'react'

import { of, BehaviorSubject, Observable, from, lastValueFrom } from 'rxjs'
import { map, concatMap } from 'rxjs/operators'

import { type ErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import type { ExternalRepoURLsResult, ExternalRepoURLsVariables } from '../../graphql-operations'

import { BlameHunkData, fetchBlameHunksMemoized } from './shared'

const IS_BLAME_VISIBLE_STORAGE_KEY = 'GitBlame.isVisible'
const isBlameVisible = new BehaviorSubject<boolean | undefined>(undefined)

export const useBlameVisibility = (isPackage: boolean): [boolean, (isVisible: boolean) => void] => {
    const [isVisibleFromLocalStorage, updateLocalStorageValue] = useLocalStorage(IS_BLAME_VISIBLE_STORAGE_KEY, false)
    const isVisibleFromObservable = useObservable(isBlameVisible)
    const setIsBlameVisible = useCallback(
        (isVisible: boolean): void => {
            isBlameVisible.next(isVisible)
            updateLocalStorageValue(isVisible)
        },
        [updateLocalStorageValue]
    )

    return [!isPackage && (isVisibleFromObservable ?? isVisibleFromLocalStorage), setIsBlameVisible]
}

/**
 * For performance reasons, the hunks array can be mutated in place. To still be
 * able to propagate updates accordingly, this is wrapped in a ref object that
 * can be recreated whenever we emit new values.
 */
export const useBlameHunks = ({
    isPackage,
    repoName,
    revision,
    filePath,
}: {
    isPackage: boolean
    repoName: string
    revision: string
    filePath: string
}): BlameHunkData | ErrorLike => {
    const [isBlameVisible] = useBlameVisibility(isPackage)
    const shouldFetchBlame = isBlameVisible

    const stream = useMemo(
        () =>
            shouldFetchBlame
                ? fetchBlameWithExternalURLs({ revision, repoName, filePath })
                : of({ current: undefined, externalURLs: undefined }),
        [shouldFetchBlame, revision, repoName, filePath]
    )
    try {
        const hunks = useObservable(stream)
        return hunks || { current: undefined, externalURLs: undefined }
    } catch (error) {
        return { message: error.toString() }
    }
}

async function fetchRepositoryData(repoName: string): Promise<Omit<BlameHunkData, 'current'>> {
    return lastValueFrom(
        requestGraphQL<ExternalRepoURLsResult, ExternalRepoURLsVariables>(
            gql`
                query ExternalRepoURLs($repo: String!) {
                    repository(name: $repo) {
                        externalURLs {
                            url
                            serviceKind
                        }
                    }
                }
            `,
            { repo: repoName }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => ({
                externalURLs: repository?.externalURLs,
            }))
        )
    )
}

function fetchBlameWithExternalURLs({
    repoName,
    revision,
    filePath,
}: {
    repoName: string
    revision: string
    filePath: string
}): Observable<BlameHunkData> {
    const blameHunks = fetchBlameHunksMemoized({ repoName, revision, filePath })
    const repositoryData = from(fetchRepositoryData(repoName))
    return repositoryData.pipe(
        concatMap(repoData =>
            blameHunks.pipe(
                map(hunks => ({
                    ...repoData,
                    current: hunks,
                }))
            )
        )
    )
}

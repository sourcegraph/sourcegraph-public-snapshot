import { ApolloClient } from '@apollo/client'
import { mdiFileDocumentOutline } from '@mdi/js'

import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { Icon } from '@sourcegraph/wildcard'

import { getWebGraphQLClient } from '../../backend/graphql'
import { SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { createUrlFunction } from '../../fuzzyFinder/WordSensitiveFuzzySearch'
import {
    FileNamesResult,
    FileNamesVariables,
    FuzzyFinderFilesResult,
    FuzzyFinderFilesVariables,
} from '../../graphql-operations'

import { FuzzyFSM, newFuzzyFSMFromValues } from './FuzzyFsm'
import { emptyFuzzyCache, PersistableQueryResult } from './FuzzyLocalCache'
import { FuzzyQuery } from './FuzzyQuery'
import { FuzzyRepoRevision, fuzzyRepoRevisionSearchFilter } from './FuzzyRepoRevision'

export const FUZZY_GIT_LSFILES_QUERY = gql`
    query FileNames($repository: String!, $commit: String!) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                id
                fileNames
            }
        }
    }
`

export const FUZZY_FILES_QUERY = gql`
    query FuzzyFinderFiles($query: String!) {
        search(patternType: regexp, query: $query) {
            results {
                results {
                    ... on FileMatch {
                        __typename
                        ...FileMatchFields
                    }
                }
            }
        }
    }
    fragment FileMatchFields on FileMatch {
        symbols {
            name
            containerName
            kind
            language
            url
        }
        repository {
            name
        }
        file {
            path
        }
    }
`

export async function loadFilesFSM(
    apolloClient: ApolloClient<object> | undefined,
    repoRevision: FuzzyRepoRevision,
    createURL: createUrlFunction
): Promise<FuzzyFSM> {
    try {
        const client = apolloClient || (await getWebGraphQLClient())
        const response = await client.query<FileNamesResult, FileNamesVariables>({
            query: getDocumentNode(FUZZY_GIT_LSFILES_QUERY),
            variables: { repository: repoRevision.repositoryName, commit: repoRevision.revision },
        })
        if (response.errors && response.errors.length > 0) {
            return { key: 'failed', errorMessage: JSON.stringify(response.errors) }
        }
        if (response.error) {
            return { key: 'failed', errorMessage: JSON.stringify(response.error) }
        }
        const filenames = response.data.repository?.commit?.fileNames || []
        return newFuzzyFSM(filenames, createURL)
    } catch (error) {
        return { key: 'failed', errorMessage: JSON.stringify(error) }
    }
}

export function newFuzzyFSM(filenames: string[], createUrl: createUrlFunction): FuzzyFSM {
    return newFuzzyFSMFromValues(
        filenames.map(file => ({
            text: file,
            icon: <Icon aria-label={file} svgPath={mdiFileDocumentOutline} />,
        })),
        createUrl
    )
}

export class FuzzyFiles extends FuzzyQuery {
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        onNamesChanged: () => void,
        private readonly repoRevision: React.MutableRefObject<FuzzyRepoRevision>,
        private readonly isGlobalFilesRef: React.MutableRefObject<boolean>
    ) {
        // Symbol results should not be cached because stale symbol data is complicated to evict/invalidate.
        super(onNamesChanged, emptyFuzzyCache)
    }

    /* override */ protected searchValues(): SearchValue[] {
        return [...this.queryResults.values()].map(({ text, url }) => ({
            text,
            url,
            icon: <Icon aria-label={text} svgPath={mdiFileDocumentOutline} />,
        }))
    }

    /* override */ protected rawQuery(query: string): string {
        const repoFilter = this.isGlobalFilesRef.current ? '' : fuzzyRepoRevisionSearchFilter(this.repoRevision.current)
        return `${repoFilter}type:path count:100 ${query}`
    }

    /* override */ protected async handleRawQueryPromise(query: string): Promise<PersistableQueryResult[]> {
        const client = this.client || (await getWebGraphQLClient())
        const response = await client.query<FuzzyFinderFilesResult, FuzzyFinderFilesVariables>({
            query: getDocumentNode(FUZZY_FILES_QUERY),
            variables: { query },
        })
        const results = response.data?.search?.results?.results || []
        const queryResults: PersistableQueryResult[] = []
        for (const result of results) {
            if (result.__typename === 'FileMatch') {
                queryResults.push({
                    text: `${result.repository.name}/${result.file.path}`,
                    url: `/${result.repository.name}/-/blob/${result.file.path}`,
                })
            }
        }
        return queryResults
    }
}

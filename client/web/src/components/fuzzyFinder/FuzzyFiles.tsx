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

function fileIcon(filename: string): JSX.Element {
    return (
        <span className="fuzzy-repos-result-icon">
            <Icon aria-label={filename} svgPath={mdiFileDocumentOutline} className="fuzzy-repos-result-icon" />
        </span>
    )
}

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
        filenames.map(text => ({
            text,
            icon: fileIcon(text),
        })),
        createUrl
    )
}

export class FuzzyRepoFiles {
    private fsm: FuzzyFSM = { key: 'empty' }
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        private readonly createURL: createUrlFunction,
        private readonly onNamesChanged: () => void,
        private readonly repoRevision: FuzzyRepoRevision
    ) {}
    public fuzzyFSM(): FuzzyFSM {
        return this.fsm
    }
    public handleQuery(): void {
        if (this.fsm.key === 'empty') {
            this.download().then(
                () => {},
                () => {}
            )
        }
    }
    private async download(): Promise<PersistableQueryResult[]> {
        const client = this.client || (await getWebGraphQLClient())
        this.fsm = { key: 'downloading' }
        const response = await client.query<FileNamesResult, FileNamesVariables>({
            query: getDocumentNode(FUZZY_GIT_LSFILES_QUERY),
            variables: {
                repository: this.repoRevision.repositoryName,
                commit: this.repoRevision.revision,
            },
        })
        const filenames = response.data.repository?.commit?.fileNames || []
        const values: SearchValue[] = filenames.map(text => ({
            text,
            icon: fileIcon(text),
        }))
        this.updateFSM(newFuzzyFSMFromValues(values, this.createURL))
        this.loopIndexing()
        return values
    }
    private updateFSM(newFSM: FuzzyFSM): void {
        this.fsm = newFSM
        this.onNamesChanged()
    }
    private loopIndexing(): void {
        if (this.fsm.key === 'indexing') {
            this.fsm.indexing.continueIndexing().then(
                next => {
                    if (next.key === 'ready') {
                        this.updateFSM({ key: 'ready', fuzzy: next.value })
                    } else {
                        this.updateFSM({ key: 'indexing', indexing: next })
                        this.loopIndexing()
                    }
                },
                error => {
                    this.updateFSM({ key: 'failed', errorMessage: JSON.stringify(error) })
                }
            )
        }
    }
}

export class FuzzyFiles extends FuzzyQuery {
    private readonly isGlobalFiles = true
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        onNamesChanged: () => void,
        private readonly repoRevision: React.MutableRefObject<FuzzyRepoRevision>
    ) {
        // Symbol results should not be cached because stale symbol data is complicated to evict/invalidate.
        super(onNamesChanged, emptyFuzzyCache)
    }

    /* override */ protected searchValues(): SearchValue[] {
        return [...this.queryResults.values()].map(({ text, url }) => ({
            text,
            url,
            icon: fileIcon(text),
        }))
    }

    /* override */ protected rawQuery(query: string): string {
        const repoFilter = this.isGlobalFiles ? '' : fuzzyRepoRevisionSearchFilter(this.repoRevision.current)
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

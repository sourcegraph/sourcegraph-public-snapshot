import type { ApolloClient } from '@apollo/client'
import { mdiFileDocumentOutline } from '@mdi/js'

import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { Icon } from '@sourcegraph/wildcard'

import { getWebGraphQLClient } from '../../backend/graphql'
import type { FuzzySearchTransformer, createUrlFunction } from '../../fuzzyFinder/FuzzySearch'
import type { SearchValue } from '../../fuzzyFinder/SearchValue'
import type {
    FileNamesResult,
    FileNamesVariables,
    FuzzyFinderFilesResult,
    FuzzyFinderFilesVariables,
} from '../../graphql-operations'
import type { UserHistory } from '../useUserHistory'

import { type FuzzyFSM, newFuzzyFSMFromValues } from './FuzzyFsm'
import { emptyFuzzyCache, type PersistableQueryResult } from './FuzzyLocalCache'
import { FuzzyQuery } from './FuzzyQuery'
import { type FuzzyRepoRevision, fuzzyRepoRevisionSearchFilter } from './FuzzyRepoRevision'

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
        <span className="mr-1">
            <Icon aria-label={filename} svgPath={mdiFileDocumentOutline} className="mr-1" />
        </span>
    )
}

export class FuzzyRepoFiles {
    private fsm: FuzzyFSM = { key: 'empty' }
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        private readonly createURL: createUrlFunction,
        private readonly onNamesChanged: () => void,
        private readonly repoRevision: FuzzyRepoRevision,
        private readonly userHistory: UserHistory
    ) {}
    public fuzzyFSM(): FuzzyFSM {
        return this.fsm
    }
    public handleQuery(): void {
        if (this.fsm.key === 'empty') {
            this.fsm = { key: 'downloading' }
            this.download().then(
                () => {},
                () => {}
            )
        }
    }
    private async download(): Promise<PersistableQueryResult[]> {
        const client = this.client || (await getWebGraphQLClient())
        const response = await client.query<FileNamesResult, FileNamesVariables>({
            query: getDocumentNode(FUZZY_GIT_LSFILES_QUERY),
            variables: {
                repository: this.repoRevision.repositoryName,
                commit: this.repoRevision.revision,
            },
        })
        const filenames = response.data.repository?.commit?.fileNames || []
        const values: SearchValue[] = filenames.map<SearchValue>(text => ({
            text,
            icon: fileIcon(text),
            historyRanking: () => this.userHistory.lastAccessedFilePath(this.repoRevision.repositoryName, text),
        }))
        this.updateFSM(newFuzzyFSMFromValues(values, { createURL: this.createURL, transformer: lineNumberTransformer }))
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

export interface FuzzyFileQuery {
    filename: string
    line?: number
    column?: number
}

export function parseFuzzyFileQuery(query: string): FuzzyFileQuery {
    if (query.endsWith(':')) {
        // Infer the line number or column number if the user is typing :.
        // Without this logic, the fuzzy finder briefly becomes empty for a
        // split second when the user types a : character before typing the line
        // number or column number.
        query = query + '0'
    }
    const lineNumberPattern = /^([^:]+):(\d+)(?::(\d+))?$/
    const match = query.match(lineNumberPattern)
    if (match && match.length > 3) {
        const value: FuzzyFileQuery = { filename: match[1] }
        const line = Number.parseInt(match[2], 10)
        if (isFinite(line)) {
            value.line = line
        }
        const column = Number.parseInt(match[3], 10)
        if (isFinite(column)) {
            value.column = column
        }
        return value
    }
    return { filename: query }
}

const lineNumberTransformer: FuzzySearchTransformer = {
    modifyQuery: (query: string) => {
        const parsed = parseFuzzyFileQuery(query)
        return parsed.filename
    },
    modifyURL: (query: string, url: string) => {
        const parsed = parseFuzzyFileQuery(query)
        if (parsed.line !== undefined && isFinite(parsed.line)) {
            return `${url}?L${parsed.line}`
        }
        return url
    },
}

export class FuzzyFiles extends FuzzyQuery {
    private readonly isGlobalFiles = true
    constructor(
        private readonly client: ApolloClient<object> | undefined,
        onNamesChanged: () => void,
        private readonly repoRevision: React.MutableRefObject<FuzzyRepoRevision>,
        private readonly userHistory: UserHistory
    ) {
        super(onNamesChanged, emptyFuzzyCache, { transformer: lineNumberTransformer })
    }

    /* override */ protected searchValues(): SearchValue[] {
        return [...this.queryResults.values()].map<SearchValue>(({ text, url, repoName, filePath }) => ({
            text,
            url,
            icon: fileIcon(text),
            historyRanking: () =>
                repoName && filePath ? this.userHistory.lastAccessedFilePath(repoName, filePath) : undefined,
            ranking: repoName ? this.userHistory.lastAccessedRepo(repoName) : undefined,
        }))
    }

    /* override */ protected rawQuery(query: string): string {
        const repoFilter = this.isGlobalFiles ? '' : fuzzyRepoRevisionSearchFilter(this.repoRevision.current)
        return `${repoFilter}type:path count:100 ${parseFuzzyFileQuery(query).filename}`
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
                    repoName: result.repository.name,
                    filePath: result.file.path,
                    text: `${result.repository.name}/${result.file.path}`,
                    url: `/${result.repository.name}/-/blob/${result.file.path}`,
                })
            }
        }
        return queryResults
    }
}

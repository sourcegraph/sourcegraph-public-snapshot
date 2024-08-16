import { differenceInHours, formatISO, parseISO } from 'date-fns'
import { lastValueFrom } from 'rxjs'
import { type Readable, readable } from 'svelte/store'
import { z } from 'zod'

import { dev } from '$app/environment'
import { isDefined } from '$lib/common'
import { streamComputeQuery } from '$lib/shared'
import { temporarySetting } from '$lib/temporarySettings'
import type { ProductStatusType } from '$lib/wildcard/ProductStatusBadge.svelte'

export interface QueryExample {
    fileName: string
    repoName: string
    orgName?: string
}

interface QueryExamplesContent {
    repositoryName: string
    filePath: string
}

const defaultQueryExamplesContent: QueryExamplesContent = {
    repositoryName: 'org/repo',
    filePath: 'file.go',
}

/**
 * Returns a store for query examples content for dotcom.
 */
export function queryExampleDotcom(): Readable<QueryExample> {
    return readable<QueryExample>({
        fileName: 'test',
        repoName: 'facebook/react',
        orgName: 'kubernetes/',
    })
}

/**
 * Returns a store for query examples content for enterprise instances.
 */
export function queryExampleEnterprise(): Readable<QueryExample | null> {
    return readable<QueryExample | null>(null, set => {
        const cachedQueryExample = temporarySetting('search.homepage.queryExamplesContent')
        cachedQueryExample
            .value()
            .then(result => {
                if (!result || hasQueryExamplesContentCacheExpired(result.lastCachedTimestamp)) {
                    return fetchComputedQueryExample().then(queryExamplesContent => {
                        cachedQueryExample.setValue({
                            ...queryExamplesContent,
                            lastCachedTimestamp: formatISO(Date.now()),
                        })
                        return queryExamplesContent
                    })
                }
                return result
            })
            .catch(error => {
                if (dev) {
                    console.error(error)
                }
                return defaultQueryExamplesContent
            })
            .then(queryExamplesContent => {
                const { repositoryName, filePath } = queryExamplesContent

                const { singleRepoExample, orgReposExample } = getRepoFilterExamples(repositoryName)
                const filePathParts = filePath.split('/')
                const fileName = quoteIfNeeded(filePathParts.at(-1)!)
                set({
                    fileName,
                    repoName: singleRepoExample,
                    orgName: orgReposExample,
                })
            })
    })
}

function quoteIfNeeded(value: string): string {
    return value.includes(' ') ? `"${value}"` : value
}

function getRepoFilterExamples(repositoryName: string): { singleRepoExample: string; orgReposExample?: string } {
    const repositoryNameParts = repositoryName.split('/')

    if (repositoryNameParts.length <= 1) {
        return { singleRepoExample: quoteIfNeeded(repositoryName) }
    }

    const repoName = repositoryNameParts.at(-1)
    const repoOrg = repositoryNameParts.at(-2)
    return {
        singleRepoExample: quoteIfNeeded(`${repoOrg}/${repoName}`),
        orgReposExample: quoteIfNeeded(`${repoOrg}/`),
    }
}

function hasQueryExamplesContentCacheExpired(lastCachedTimestamp: string): boolean {
    return differenceInHours(Date.now(), parseISO(lastCachedTimestamp)) > 24
}

const ComputeResultSchema = z.array(
    z.object({
        kind: z.string(),
        value: z.string(),
    })
)

async function fetchComputedQueryExample(): Promise<QueryExamplesContent> {
    const results = await lastValueFrom(streamComputeQuery(`type:diff count:1 content:output((.|\n)* -> $repo,|$path)`))
    const firstComputeOutput = results
        .map(result => JSON.parse(result))
        .filter(result => Array.isArray(result))
        .flatMap(result => ComputeResultSchema.parse(result))
        .find(result => result.kind === 'output')

    return firstComputeOutput
        ? getQueryExamplesContentFromComputeOutput(firstComputeOutput.value)
        : defaultQueryExamplesContent
}

function getQueryExamplesContentFromComputeOutput(computeOutput: string): QueryExamplesContent {
    const [repositoryName, filePath] = computeOutput.trim().split(',|')
    return {
        repositoryName,
        filePath,
    }
}

export interface QueryExamplesSection {
    title: string
    queryExamples: {
        query: string
        helperText?: string
        productStatus?: ProductStatusType
    }[]
}

export const exampleQueryColumns: QueryExamplesSection[][] = [
    [
        {
            title: 'Find usage examples',
            queryExamples: [
                { query: 'context.WithCancel lang:go' },
                { query: '<Suspense lang:typescript' },
                { query: 'readFileSync lang:javascript' },
                { query: 'import torch lang:python' },
            ],
        },
    ],
    [
        {
            title: 'Find TODOs in a repository',
            queryExamples: [{ query: 'repo:facebook/react TODO' }],
        },
        {
            title: 'See API usage and changes over time',
            queryExamples: [{ query: 'repo:pytorch/pytorch type:diff is_cpu' }],
        },
    ],
]

export function getKeywordExamples(
    fileName: string,
    singleRepoExample: string,
    orgReposExample: string | undefined
): QueryExamplesSection[][] {
    return [
        [
            {
                title: 'Search in files, paths, and repository-names',
                queryExamples: [
                    { query: 'test server', helperText: '(both terms anywhere)', productStatus: 'new' },
                    { query: '"Error 1001"', helperText: '(specific string)', productStatus: 'new' },
                    {
                        query: '"\\"Error 1001\\""',
                        helperText: '(specific string containing quotes)',
                        productStatus: 'new',
                    },
                    { query: 'foo OR bar' },
                    { query: '/open(File|Dir)/', helperText: '(regular expression)' },
                ],
            },
            {
                title: 'Search in commit diffs',
                queryExamples: [{ query: 'type:diff after:1week fix' }, { query: 'type:diff author:alice add' }],
            },
        ],
        [
            {
                title: 'Filter by...',
                queryExamples: [
                    { query: `file:${fileName} foo` },
                    { query: `repo:${singleRepoExample}` },
                    orgReposExample
                        ? { query: `repo:${orgReposExample}`, helperText: '(all repositories in org)' }
                        : null,
                    { query: 'lang:javascript' },
                ].filter(isDefined),
            },
            {
                title: 'Advanced',
                queryExamples: [
                    { query: 'repo:has.description(foo)' },
                    { query: 'file:^some_path file:has.owner(alice)' },
                    { query: 'file:^some_path select:file.owners' },
                    { query: 'file:has.commit.after(1week)' },
                ],
            },
        ],
    ]
}

export function getStandardExamples(
    fileName: string,
    singleRepoExample: string,
    orgReposExample: string | undefined
): QueryExamplesSection[][] {
    return [
        [
            {
                title: 'Search in files',
                queryExamples: [
                    { query: 'fetch(' },
                    { query: 'some error message', helperText: '(no quotes needed)' },
                    { query: 'foo AND bar' },
                    { query: '/open(File|Dir)/', helperText: '(regular expression)' },
                ],
            },
            {
                title: 'Search in commit diffs',
                queryExamples: [{ query: 'type:diff after:1week fix' }, { query: 'type:diff author:alice add' }],
            },
        ],
        [
            {
                title: 'Filter by...',
                queryExamples: [
                    { query: `file:${fileName} foo` },
                    { query: `repo:${singleRepoExample}` },
                    orgReposExample
                        ? { query: `repo:${orgReposExample}`, helperText: '(all repositories in org)' }
                        : null,
                    { query: 'lang:javascript' },
                ].filter(isDefined),
            },
            {
                title: 'Advanced',
                queryExamples: [
                    { query: 'repo:has.description(foo)' },
                    { query: 'file:^some_path file:has.owner(alice)' },
                    { query: 'file:^some_path select:file.owners' },
                    { query: 'file:has.commit.after(1week)' },
                ],
            },
        ],
    ]
}

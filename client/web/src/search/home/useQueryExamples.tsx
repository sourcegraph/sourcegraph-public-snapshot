import React, { useMemo, useEffect, useState } from 'react'

import { mdiOpenInNew } from '@mdi/js'
import { differenceInHours, formatISO, parseISO } from 'date-fns'

import { streamComputeQuery } from '@sourcegraph/shared/src/search/stream'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Icon, Link } from '@sourcegraph/wildcard'

export interface QueryExamplesContent {
    repositoryName: string
    filePath: string
    author: string
}

interface QueryExample {
    title: string
    queryExamples: {
        id: string
        query: string
        helperText?: string
    }[]
    footer?: React.ReactElement
}

interface ComputeResult {
    kind: string
    value: string
}

const defaultQueryExamplesContent = {
    repositoryName: 'organization/repo-name',
    author: 'Logan Smith',
    filePath: 'filename.go',
}

function hasQueryExamplesContentCacheExpired(lastCachedTimestamp: string): boolean {
    return differenceInHours(Date.now(), parseISO(lastCachedTimestamp)) > 24
}

function quoteIfNeeded(value: string): string {
    return value.includes(' ') ? `"${value}"` : value
}

function getQueryExamplesContentFromComputeOutput(computeOutput: string): QueryExamplesContent {
    const [repositoryName, author, filePath] = computeOutput.trim().split(',|')
    return {
        repositoryName,
        filePath,
        author,
    }
}

function getRepoFilterExamples(repositoryName: string): { singleRepoExample: string; orgReposExample?: string } {
    const repositoryNameParts = repositoryName.split('/')

    if (repositoryNameParts.length <= 1) {
        return { singleRepoExample: quoteIfNeeded(repositoryName) }
    }

    const repoName = repositoryNameParts[repositoryNameParts.length - 1]
    const repoOrg = repositoryNameParts[repositoryNameParts.length - 2]
    return {
        singleRepoExample: quoteIfNeeded(`${repoOrg}/${repoName}`),
        orgReposExample: quoteIfNeeded(`${repoOrg}/.*`),
    }
}

export function useQueryExamples(): QueryExample[][] {
    const [queryExamplesContent, setQueryExamplesContent] = useState<QueryExamplesContent>()
    const [
        cachedQueryExamplesContent,
        setCachedQueryExamplesContent,
        cachedQueryExamplesContentLoadStatus,
    ] = useTemporarySetting('search.homepage.queryExamplesContent')

    useEffect(() => {
        if (queryExamplesContent || cachedQueryExamplesContentLoadStatus === 'initial') {
            return
        }

        if (
            cachedQueryExamplesContentLoadStatus === 'loaded' &&
            cachedQueryExamplesContent &&
            !hasQueryExamplesContentCacheExpired(cachedQueryExamplesContent.lastCachedTimestamp)
        ) {
            setQueryExamplesContent(cachedQueryExamplesContent)
            return
        }

        // We are using `,|` as the separator so we can "safely" split the compute output.
        const subscription = streamComputeQuery(
            'type:diff count:1 content:output((.|\n)* -> $repo,|$author,|$path)'
        ).subscribe(
            results => {
                const firstComputeOutput = results
                    .flatMap(result => JSON.parse(result) as ComputeResult)
                    .find(result => result.kind === 'output')

                const queryExamplesContent = firstComputeOutput
                    ? getQueryExamplesContentFromComputeOutput(firstComputeOutput.value)
                    : defaultQueryExamplesContent

                setQueryExamplesContent(queryExamplesContent)
                setCachedQueryExamplesContent({
                    ...queryExamplesContent,
                    lastCachedTimestamp: formatISO(Date.now()),
                })
            },
            () => {
                // In case of an error set default content.
                setQueryExamplesContent(defaultQueryExamplesContent)
            }
        )

        return () => subscription.unsubscribe()
    }, [
        queryExamplesContent,
        cachedQueryExamplesContent,
        setCachedQueryExamplesContent,
        cachedQueryExamplesContentLoadStatus,
    ])

    return useMemo(() => {
        if (!queryExamplesContent) {
            return []
        }
        const { repositoryName, filePath, author } = queryExamplesContent

        const { singleRepoExample, orgReposExample } = getRepoFilterExamples(repositoryName)
        const filePathParts = filePath.split('/')
        const fileName = quoteIfNeeded(filePathParts[filePathParts.length - 1])
        const quotedAuthor = quoteIfNeeded(author)

        return [
            [
                {
                    title: 'Scope search to specific repos',
                    queryExamples: [
                        { id: 'single-repo', query: `repo:${singleRepoExample}` },
                        { id: 'org-repos', query: orgReposExample ? `repo:${orgReposExample}` : '' },
                    ],
                },
                {
                    title: 'Jump into code navigation',
                    queryExamples: [
                        { id: 'file-filter', query: `file:${fileName}` },
                        { id: 'type-symbol', query: 'type:symbol SymbolName' },
                    ],
                },
                {
                    title: 'Explore code history',
                    queryExamples: [
                        { id: 'type-diff-author', query: `type:diff author:${quotedAuthor}` },
                        { id: 'type-commit-message', query: 'type:commit some message' },
                        { id: 'type-diff-after', query: 'type:diff after:"1 year ago"' },
                    ],
                },
            ],
            [
                {
                    title: 'Find content or patterns',
                    queryExamples: [
                        { id: 'exact-matches', query: 'some exact error message', helperText: 'No quotes needed' },
                        { id: 'regex-pattern', query: '/regex.*pattern/' },
                    ],
                },
                {
                    title: 'Get logical',
                    queryExamples: [
                        { id: 'or-operator', query: 'lang:javascript OR lang:typescript' },
                        { id: 'and-operator', query: 'hello AND world' },
                        { id: 'not-operator', query: 'lang:go NOT file:main.go' },
                    ],
                },
                {
                    title: 'Get advanced',
                    queryExamples: [{ id: 'repo-has-description', query: 'repo:has.description(hello world)' }],
                    footer: (
                        <small className="d-block mt-3">
                            <Link target="blank" to="/help/code_search/reference/queries">
                                Complete query reference{' '}
                                <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                            </Link>
                        </small>
                    ),
                },
            ],
        ]
    }, [queryExamplesContent])
}

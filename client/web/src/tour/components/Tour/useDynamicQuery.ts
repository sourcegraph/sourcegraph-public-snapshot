import { useContext, useEffect, useMemo, useState } from 'react'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'

import { defaultSnippets } from '../../data'

import { TourContext } from './context'
import { QueryPlaceholder, containsPlaceholder, isNotNullOrUndefined } from './utils'

export function useDynamicQuery(template: string, snippets?: string[] | Record<string, string[]>): string {
    const [query, setQuery] = useState<string>('')

    const hasSnippet = containsPlaceholder(template, QueryPlaceholder.Snippet)
    const { userInfo, isQuerySuccessful } = useContext(TourContext)
    const { repo = '', language = '', email = '' } = userInfo ?? {}

    const baseQuery = useMemo(
        () =>
            buildQuery(template, {
                [QueryPlaceholder.Repo]: displayRepoName(repo),
                [QueryPlaceholder.Lang]: language,
                [QueryPlaceholder.Email]: email,
            }),
        [repo, language, email, template]
    )

    useEffect(() => {
        if (hasSnippet) {
            // We have multiple snippets available:
            // - Specific snippets defined in the step (snippets prop)
            // - Language specific default snippets
            // - Generic default snippets
            // We try each set (from specific to generic) and use the first one that is successful.

            const snippetsQueue = [
                // Configured snippets (if any)
                snippets
                    ? Array.isArray(snippets)
                        ? snippets
                        : language
                        ? getLanguageSnippets(snippets, language)
                        : null
                    : null,
                // Default language snippets
                language ? getLanguageSnippets(defaultSnippets, language) : null,
                // Default generic snippets
                defaultSnippets['*'],
            ].filter(isNotNullOrUndefined)

            findQueryFromQueue(baseQuery, snippetsQueue, isQuerySuccessful).then(setQuery, () =>
                // fall back to using an empty snippets in the query
                setQuery(buildQuery(baseQuery, { [QueryPlaceholder.Snippet]: '' }))
            )
        } else {
            // fall back to using an empty snippets in the query
            setQuery(buildQuery(baseQuery, { [QueryPlaceholder.Snippet]: '' }))
        }
    }, [baseQuery, hasSnippet, snippets, language, isQuerySuccessful])

    return query
}

/**
 * Replaces '$$abc' variables in a query template with the corresponding value from the
 * `variables` map.
 */
function buildQuery(template: string, variables: Record<string, string>): string {
    return template.replaceAll(/\$\$\w+/g, match => variables[match] ?? match)
}

/**
 * Helper function for accessing a language -> snippets map. The language names are
 * compared ignoring case.
 */
function getLanguageSnippets(snippets: Record<string, string[]>, language: string): string[] | null {
    const languageLower = language.toLowerCase()
    for (const [langKey, values] of Object.entries(snippets)) {
        if (langKey.toLowerCase() === languageLower) {
            return values
        }
    }
    return null
}

function findQuery(
    baseQuery: string,
    snippets: string[],
    isQuerySuccessful: (query: string) => Promise<boolean>
): Promise<string> {
    const promises = []
    for (const snippet of snippets) {
        const query = buildQuery(baseQuery, { [QueryPlaceholder.Snippet]: snippet })
        promises.push(
            isQuerySuccessful(query).then(
                // The rejection reason isn not relevant because we use Promise.any below to get the first
                // resolved promise.
                isSuccessful => (isSuccessful ? query : Promise.reject(new Error('not successful')))
            )
        )
    }

    return Promise.any(promises)
}

async function findQueryFromQueue(
    query: string,
    queue: string[][],
    isQuerySuccessful: (query: string) => Promise<boolean>
): Promise<string> {
    for (const next of queue) {
        try {
            return await findQuery(query, next, isQuerySuccessful)
        } catch {
            // try next in queue
        }
    }
    throw new Error('Unable to determine query that produces results')
}

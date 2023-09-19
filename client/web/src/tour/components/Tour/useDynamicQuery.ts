import { useContext, useEffect, useMemo, useState } from 'react'

import { defaultSnippets } from '../../data'

import { TourContext } from './context'

export function useDynamicQuery(template: string, snippets?: string[] | Record<string, string[]>): string {
    const [query, setQuery] = useState<string>('')

    const hasSnippet = hasSnippetPlaceholder(template)
    const { userConfig, isQuerySuccessful } = useContext(TourContext)
    const { userorg, userrepo, userlang } = userConfig ?? {}

    const baseQuery = useMemo(
        () =>
            userorg && userrepo && userlang
                ? buildQuery(template, {
                      [QueryPlaceholder.Org]: userorg,
                      [QueryPlaceholder.Repo]: userrepo,
                      [QueryPlaceholder.Lang]: userlang,
                  })
                : null,
        [userorg, userrepo, userlang, template]
    )

    useEffect(() => {
        if (baseQuery && hasSnippet) {
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
                        : userlang
                        ? getLanguageSnippets(snippets, userlang)
                        : null
                    : null,
                // Default language snippets
                userlang ? getLanguageSnippets(defaultSnippets, userlang) : null,
                // Default generic snippets
                defaultSnippets['*'],
            ].filter((snippets): snippets is string[] => snippets !== null)

            findQueryFromQueue(baseQuery, snippetsQueue, isQuerySuccessful).then(setQuery, () =>
                // fall back to using an empty snippets in the query
                setQuery(buildQuery(baseQuery, { [QueryPlaceholder.Snippet]: '' }))
            )
        } else if (baseQuery) {
            // fall back to using an empty snippets in the query
            setQuery(buildQuery(baseQuery, { [QueryPlaceholder.Snippet]: '' }))
        }
    }, [baseQuery, hasSnippet, snippets, userlang, isQuerySuccessful])

    return query
}

enum QueryPlaceholder {
    Snippet = '$$snippet',
    Org = '$$userorg',
    Repo = '$$userrepo',
    Lang = '$$userlang',
}

function hasSnippetPlaceholder(queryTemplate: string): boolean {
    return queryTemplate.includes(QueryPlaceholder.Snippet)
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

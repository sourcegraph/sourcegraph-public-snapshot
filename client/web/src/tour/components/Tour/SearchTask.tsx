import { type FC, useState, useEffect, useMemo } from 'react'

import classNames from 'classnames'
import { noop, memoize } from 'lodash'
import { Subscriber, Subscription, fromEvent, of } from 'rxjs'
import { map } from 'rxjs/operators'

import {
    LATEST_VERSION,
    MessageHandlers,
    SearchMatch,
    messageHandlers,
    search,
    switchAggregateSearchResults,
    observeMessages,
    SearchEvent,
} from '@sourcegraph/shared/src/search/stream'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Link } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../graphql-operations'
import { defaultSnippets } from '../../data'

interface SearchTaskProps {
    label: string
    template: string
    snippets?: string[] | Record<string, string[]>
    handleLinkClick: (event: React.MouseEvent<HTMLElement, MouseEvent> | React.KeyboardEvent<HTMLElement>) => void
}

export const SearchTask: FC<SearchTaskProps> = ({ template, snippets, label, handleLinkClick }) => {
    const [selectedQuery, setSelectedQuery] = useState<string>('')

    const [org] = useState<string>('sourcegraph')
    const [repo] = useState<string>('sourcegraph')
    const [lang] = useState<string>('go')

    const hasSnippet = hasSnippetPlaceholder(template)
    const baseQuery = useMemo(
        () =>
            org && repo && lang
                ? buildQuery(template, {
                      [QueryPlacholder.Org]: org,
                      [QueryPlacholder.Repo]: repo,
                      [QueryPlacholder.Lang]: lang,
                  })
                : null,
        [org, repo, lang]
    )

    useEffect(() => {
        if (baseQuery && hasSnippet && lang) {
            const snippetsQueue = [
                // Configures snippets (if any)
                snippets ? (Array.isArray(snippets) ? snippets : getLanguageSnippets(snippets, lang)) : null,
                // Hardcoded default language snippets
                getLanguageSnippets(defaultSnippets, lang),
                // Hardcoded default snippets for all languages
                defaultSnippets['*'],
            ].filter((snippets): snippets is string[] => snippets !== null)

            findQueryFromQueue(baseQuery, snippetsQueue).then(setSelectedQuery, () =>
                setSelectedQuery(buildQuery(baseQuery, { [QueryPlacholder.Snippet]: '' }))
            )
        } else if (baseQuery) {
            setSelectedQuery(buildQuery(baseQuery, { [QueryPlacholder.Snippet]: '' }))
        }
    }, [baseQuery, hasSnippet, snippets, lang])

    return selectedQuery ? (
        <Link
            className={classNames('flex-grow-1')}
            to={`/search?${buildSearchURLQuery(selectedQuery, SearchPatternType.standard, false)}`}
            onClick={handleLinkClick}
        >
            {label}
        </Link>
    ) : null
}

function getLanguageSnippets(snippets: Record<string, string[]>, language: string): string[] | null {
    const languageLower = language.toLowerCase()
    for (const [langKey, values] of Object.entries(snippets)) {
        if (langKey.toLowerCase() === languageLower) {
            return values
        }
    }
    return null
}

function findQuery(baseQuery: string, snippets: string[]): Promise<string> {
    let promises = []
    for (const snippet of snippets) {
        const query = buildQuery(baseQuery, { [QueryPlacholder.Snippet]: snippet })
        promises.push(isQuerySuccessful(query).then(isSuccessful => (isSuccessful ? query : Promise.reject())))
    }

    return Promise.any(promises)
}

async function findQueryFromQueue(query: string, queue: string[][]): Promise<string> {
    for (const next of queue) {
        try {
            return await findQuery(query, next)
        } catch {}
    }
    throw new Error('Unable to determine query that produces results')
}

enum QueryPlacholder {
    Snippet = '$$snippet',
    Org = '$$userorg',
    Repo = '$$userrepo',
    Lang = '$$userlang',
}

function hasSnippetPlaceholder(queryTemplate: string): boolean {
    return queryTemplate.includes(QueryPlacholder.Snippet)
}

/**
 * Replaces '$$abc' variables in a query template with the corresponding value from the
 * `variables` map.
 */
function buildQuery(template: string, variables: Record<string, string>): string {
    return template.replaceAll(/\$\$\w+/g, match => {
        return variables[match] ?? match
    })
}

/**
 * Executes the a "restricted" version of the provided query to determine whether
 * the query returns results or not.
 */
function isQuerySuccessful(query: string): Promise<boolean> {
    let dynamicQuery = `${query} timeout:3s count:1`

    if (!query.includes('type:')) {
        dynamicQuery += ' select:content'
    }

    return fetchStreamSuggestions(dynamicQuery)
        .then(results => results.length > 0)
        .catch(() => false)
}

const noopHandler = <T extends SearchEvent>(
    type: T['type'],
    eventSource: EventSource,
    _observer: Subscriber<SearchEvent>
): Subscription => fromEvent(eventSource, type).subscribe(noop)

const firstMatchMessageHandlers: MessageHandlers = {
    ...messageHandlers,
    matches: (type, eventSource, observer) =>
        observeMessages(type, eventSource).subscribe(data => {
            observer.next(data)
            // Once we observer the first `matches` event, complete the stream and close the event source.
            observer.complete()
            eventSource.close()
        }),
    progress: noopHandler,
    filters: noopHandler,
    alert: noopHandler,
}

/**
 * Initiates a streaming search, stop at the first `matches` event, and aggregate the results.
 */
const fetchStreamSuggestions = memoize(
    (query: string, sourcegraphURL?: string): Promise<SearchMatch[]> => {
        return search(
            of(query),
            {
                version: LATEST_VERSION,
                patternType: SearchPatternType.standard,
                caseSensitive: false,
                trace: undefined,
                sourcegraphURL,
            },
            firstMatchMessageHandlers
        )
            .pipe(
                switchAggregateSearchResults,
                map(suggestions => suggestions.results)
            )
            .toPromise()
    },
    (query, sourcegraphURL) => `${query} + ${sourcegraphURL}`
)

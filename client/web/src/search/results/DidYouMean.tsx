import React, { useEffect, useMemo } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ALL_LANGUAGES } from '@sourcegraph/shared/src/search/query/languageFilter'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { createLiteral, Pattern, Token } from '@sourcegraph/shared/src/search/query/token'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { CaseSensitivityProps, ParsedSearchQueryProps, PatternTypeProps, SearchContextProps } from '..'
import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'

import styles from './DidYouMean.module.scss'

// Only consider queries that have at most this many terms
const MAX_TERMS = 4

const normalizedLanguages = new Map(ALL_LANGUAGES.map(lang => [lang.toLowerCase(), lang]))

function isPattern(token: Token): token is Pattern {
    return token.type === 'pattern'
}

function matchesLanguage(token: Pattern): { success: false } | { success: true; language: string; token: Pattern } {
    const normalizedSearchTerm = token.value.toLowerCase()
    if (normalizedLanguages.has(normalizedSearchTerm)) {
        return {
            success: true,
            language: normalizedLanguages.get(normalizedSearchTerm)!,
            token,
        }
    }

    return { success: false }
}

interface Suggestion {
    type: 'languageFilter'
    query: string
    text: React.ReactElement
}

function getQuerySuggestions(query: string, patternType: SearchPatternType): Suggestion[] {
    const result: Suggestion[] = []

    const scanResult = scanSearchQuery(query, false, patternType)
    if (scanResult.type !== 'success') {
        return result
    }

    // This is used later to reconstruct the query
    const tokensWithoutContext = scanResult.term.filter(term => {
        switch (term.type) {
            case 'filter':
                if (term.field.value === 'context') {
                    return false
                }
                return true
            default:
                return true
        }
    })

    // This is used to analyse the query
    const tokensWithoutWhitespace = tokensWithoutContext.filter(term => {
        switch (term.type) {
            case 'comment':
            case 'whitespace':
                return false
            default:
                return true
        }
    })

    // Only consider queries that don't contain filters
    if (
        tokensWithoutWhitespace.length < 2 ||
        tokensWithoutWhitespace.length > MAX_TERMS ||
        !tokensWithoutWhitespace.every(isPattern)
    ) {
        return result
    }

    let matchResult = matchesLanguage(tokensWithoutWhitespace[0])
    if (!matchResult.success) {
        matchResult = matchesLanguage(tokensWithoutWhitespace[tokensWithoutWhitespace.length - 1])
    }

    if (matchResult.success) {
        const { token: matchedToken, language } = matchResult
        const updatedQuery: Token[] = tokensWithoutContext.map(
            (token: Token): Token =>
                token === matchedToken
                    ? {
                          type: 'filter',
                          field: createLiteral('lang', { start: 0, end: 0 }),
                          value: createLiteral(token.value, { start: 0, end: 0 }),
                          negated: false,
                          range: { start: 0, end: 0 },
                      }
                    : token
        )
        result.push({
            type: 'languageFilter',
            query: stringHuman(updatedQuery),
            text: (
                <span>
                    Search in <em>{language}</em> files
                </span>
            ),
        })
    }
    return result
}

interface DidYouMeanProps
    extends Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        TelemetryProps {}

export const DidYouMean: React.FunctionComponent<DidYouMeanProps> = ({
    telemetryService,
    parsedSearchQuery,
    patternType,
    caseSensitive,
    selectedSearchContextSpec,
}) => {
    const suggestions = useMemo(() => getQuerySuggestions(parsedSearchQuery, patternType), [
        parsedSearchQuery,
        patternType,
    ])

    useEffect(() => {
        if (suggestions.length > 0) {
            telemetryService.log('SearchDidYouMeanDisplayed')
        }
    }, [suggestions, telemetryService])

    if (suggestions.length > 0) {
        return (
            <div className={styles.root}>
                <h3>Did you mean:</h3>
                <ul className={styles.container}>
                    {suggestions.map(suggestion => {
                        const builtURLQuery = buildSearchURLQuery(
                            suggestion.query,
                            patternType,
                            caseSensitive,
                            selectedSearchContextSpec
                        )
                        return (
                            <li key={suggestion.query}>
                                <Link
                                    onClick={() =>
                                        telemetryService.log('SearchDidYouMeanClicked', { type: suggestion.type })
                                    }
                                    to={{ pathname: '/search', search: builtURLQuery }}
                                >
                                    <span className={styles.suggestion}>
                                        <SyntaxHighlightedSearchQuery query={suggestion.query} />
                                    </span>
                                    {suggestion.text}
                                </Link>
                            </li>
                        )
                    })}
                </ul>
            </div>
        )
    }
    return null
}

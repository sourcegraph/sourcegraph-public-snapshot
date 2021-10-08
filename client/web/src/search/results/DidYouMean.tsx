import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ALL_LANGUAGES } from '@sourcegraph/shared/src/search/query/languageFilter'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { CaseSensitivityProps, ParsedSearchQueryProps, PatternTypeProps, SearchContextProps } from '..'
import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'

import styles from './DidYouMean.module.scss'

interface DidYouMeanProps
    extends Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        Pick<VersionContextProps, 'versionContext'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {}

const normalizedLanguages = new Map(ALL_LANGUAGES.map(lang => [lang.toLowerCase(), lang]))

interface Suggestion {
    query: string
    text: React.ReactElement
}

function getQuerySuggestions(query: string, patternType: SearchPatternType): Suggestion[] {
    const result: Suggestion[] = []

    const scanResult = scanSearchQuery(query, false, patternType)
    if (scanResult.type !== 'success') {
        return result
    }

    const terms = scanResult.term.filter(term => {
        switch (term.type) {
            case 'comment':
            case 'whitespace':
                return false
            case 'filter':
                if (term.field.value === 'context') {
                    return false
                }
                return true
            default:
                return true
        }
    })

    const nTerms = terms.length
    // Query must contain 2 - 3 patterns
    if (nTerms === 1 || nTerms > 3 || !terms.every(term => term.type === 'pattern')) {
        return result
    }
    if (terms[0].type === 'pattern') {
        const normalizedSearchTerm = terms[0].value.toLowerCase()
        if (normalizedLanguages.has(normalizedSearchTerm)) {
            const queryTail = stringHuman(terms.slice(1))
            result.push({
                query: `lang:${terms[0].value} ${queryTail}`,
                text: (
                    <span>
                        Search in <em>{normalizedLanguages.get(normalizedSearchTerm)}</em> files
                    </span>
                ),
            })
        }
    }
    return result
}

export const DidYouMean: React.FunctionComponent<DidYouMeanProps> = ({
    parsedSearchQuery,
    patternType,
    caseSensitive,
    versionContext,
    selectedSearchContextSpec,
}) => {
    const suggestions = getQuerySuggestions(parsedSearchQuery, patternType)
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
                            versionContext,
                            selectedSearchContextSpec
                        )
                        return (
                            <li key={suggestion.query}>
                                <Link to={{ pathname: '/search', search: builtURLQuery }}>
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

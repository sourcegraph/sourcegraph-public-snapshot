import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { SearchStatsLanguages, summarizeSearchResultsStatsLanguages } from './SearchStatsLanguages'

describe('SearchStatsLanguages', () => {
    test('renders', () => {
        const component = render(
            <MemoryRouter>
                <CompatRouter>
                    <SearchStatsLanguages
                        query="abc"
                        stats={{
                            __typename: 'SearchResultsStats',
                            languages: [
                                { __typename: 'LanguageStatistics', name: 'A', totalLines: 100 },
                                { __typename: 'LanguageStatistics', name: 'B', totalLines: 50 },
                                { __typename: 'LanguageStatistics', name: 'C', totalLines: 10 },
                                { __typename: 'LanguageStatistics', name: 'D', totalLines: 5 },
                                { __typename: 'LanguageStatistics', name: '', totalLines: 35 },
                            ],
                        }}
                    />
                </CompatRouter>
            </MemoryRouter>
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
})

describe('summarizeSearchResultsStats', () => {
    test('collapses low-ranking entries to Other', () =>
        expect(
            summarizeSearchResultsStatsLanguages(
                [
                    { __typename: 'LanguageStatistics', name: 'A', totalLines: 100 },
                    { __typename: 'LanguageStatistics', name: 'B', totalLines: 50 },
                    { __typename: 'LanguageStatistics', name: 'C', totalLines: 10 },
                    { __typename: 'LanguageStatistics', name: 'D', totalLines: 5 },
                    { __typename: 'LanguageStatistics', name: '', totalLines: 35 },
                ],
                0.1
            )
        ).toEqual([
            { __typename: 'LanguageStatistics', name: 'A', totalLines: 100 },
            { __typename: 'LanguageStatistics', name: 'B', totalLines: 50 },
            { __typename: 'LanguageStatistics', name: '', totalLines: 35 },
            { __typename: 'LanguageStatistics', name: 'Other', totalLines: 15 },
        ]))
})

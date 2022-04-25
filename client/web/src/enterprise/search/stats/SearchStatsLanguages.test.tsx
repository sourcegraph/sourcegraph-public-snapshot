import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'

import { SearchStatsLanguages, summarizeSearchResultsStatsLanguages } from './SearchStatsLanguages'

describe('SearchStatsLanguages', () => {
    test('renders', () => {
        const component = render(
            <MemoryRouter>
                <SearchStatsLanguages
                    query="abc"
                    stats={{
                        __typename: 'SearchResultsStats',
                        approximateResultCount: '123',
                        sparkline: [],
                        languages: [
                            { __typename: 'LanguageStatistics', name: 'A', totalBytes: 0, totalLines: 100 },
                            { __typename: 'LanguageStatistics', name: 'B', totalBytes: 0, totalLines: 50 },
                            { __typename: 'LanguageStatistics', name: 'C', totalBytes: 0, totalLines: 10 },
                            { __typename: 'LanguageStatistics', name: 'D', totalBytes: 0, totalLines: 5 },
                            { __typename: 'LanguageStatistics', name: '', totalBytes: 0, totalLines: 35 },
                        ],
                    }}
                />
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
                    { __typename: 'LanguageStatistics', name: 'A', totalBytes: 0, totalLines: 100 },
                    { __typename: 'LanguageStatistics', name: 'B', totalBytes: 0, totalLines: 50 },
                    { __typename: 'LanguageStatistics', name: 'C', totalBytes: 0, totalLines: 10 },
                    { __typename: 'LanguageStatistics', name: 'D', totalBytes: 0, totalLines: 5 },
                    { __typename: 'LanguageStatistics', name: '', totalBytes: 0, totalLines: 35 },
                ],
                0.1
            )
        ).toEqual([
            { __typename: 'LanguageStatistics', name: 'A', totalBytes: 0, totalLines: 100 },
            { __typename: 'LanguageStatistics', name: 'B', totalBytes: 0, totalLines: 50 },
            { __typename: 'LanguageStatistics', name: '', totalBytes: 0, totalLines: 35 },
            { __typename: 'LanguageStatistics', name: 'Other', totalBytes: 0, totalLines: 15 },
        ]))
})

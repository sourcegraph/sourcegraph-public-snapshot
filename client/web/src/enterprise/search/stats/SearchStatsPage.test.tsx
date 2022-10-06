import { render, act } from '@testing-library/react'
import * as H from 'history'
import { MemoryRouter } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'
import { of } from 'rxjs'

import { SearchResultsStatsFields } from '../../../graphql-operations'

import { SearchStatsPage } from './SearchStatsPage'

describe('SearchStatsPage', () => {
    test('renders', () => {
        const component = render(
            <MemoryRouter>
                <CompatRouter>
                    <SearchStatsPage
                        location={H.createLocation({ pathname: '/stats', search: 'q=abc' })}
                        history={H.createMemoryHistory()}
                        _querySearchResultsStats={() =>
                            of<SearchResultsStatsFields & { limitHit: boolean }>({
                                __typename: 'SearchResultsStats',
                                languages: [
                                    { __typename: 'LanguageStatistics', name: 'A', totalLines: 100 },
                                    { __typename: 'LanguageStatistics', name: 'B', totalLines: 50 },
                                    { __typename: 'LanguageStatistics', name: 'C', totalLines: 10 },
                                    { __typename: 'LanguageStatistics', name: 'D', totalLines: 5 },
                                    { __typename: 'LanguageStatistics', name: '', totalLines: 35 },
                                ],
                                limitHit: false,
                            })
                        }
                    />
                </CompatRouter>
            </MemoryRouter>
        )
        act(() => undefined) // wait for _querySearchResultsStats to emit
        expect(component.asFragment()).toMatchSnapshot()
    })

    test('limitHit', () => {
        const component = render(
            <MemoryRouter>
                <CompatRouter>
                    <SearchStatsPage
                        location={H.createLocation({ pathname: '/stats', search: 'q=abc' })}
                        history={H.createMemoryHistory()}
                        _querySearchResultsStats={() =>
                            of<SearchResultsStatsFields & { limitHit: boolean }>({
                                __typename: 'SearchResultsStats',
                                languages: [{ __typename: 'LanguageStatistics', name: 'A', totalLines: 100 }],
                                limitHit: true,
                            })
                        }
                    />
                </CompatRouter>
            </MemoryRouter>
        )
        act(() => undefined) // wait for _querySearchResultsStats to emit
        expect(component.asFragment()).toMatchSnapshot()
    })
})

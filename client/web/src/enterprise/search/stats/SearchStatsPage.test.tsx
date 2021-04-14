import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer, { act } from 'react-test-renderer'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { SearchStatsPage } from './SearchStatsPage'

describe('SearchStatsPage', () => {
    test('renders', () => {
        const component = renderer.create(
            <MemoryRouter>
                <SearchStatsPage
                    location={H.createLocation({ pathname: '/stats', search: 'q=abc' })}
                    history={H.createMemoryHistory()}
                    _querySearchResultsStats={() =>
                        of<GQL.ISearchResultsStats & { limitHit: boolean }>({
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
                            limitHit: false,
                        })
                    }
                />
            </MemoryRouter>,
            {
                createNodeMock: () => ({ parentElement: document.implementation.createHTMLDocument().body }),
            }
        )
        act(() => undefined) // wait for _querySearchResultsStats to emit
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('limitHit', () => {
        const component = renderer.create(
            <MemoryRouter>
                <SearchStatsPage
                    location={H.createLocation({ pathname: '/stats', search: 'q=abc' })}
                    history={H.createMemoryHistory()}
                    _querySearchResultsStats={() =>
                        of<GQL.ISearchResultsStats & { limitHit: boolean }>({
                            __typename: 'SearchResultsStats',
                            approximateResultCount: '123',
                            sparkline: [],
                            languages: [
                                { __typename: 'LanguageStatistics', name: 'A', totalBytes: 0, totalLines: 100 },
                            ],
                            limitHit: true,
                        })
                    }
                />
            </MemoryRouter>,
            {
                createNodeMock: () => ({ parentElement: document.implementation.createHTMLDocument().body }),
            }
        )
        act(() => undefined) // wait for _querySearchResultsStats to emit
        expect(component.toJSON()).toMatchSnapshot()
    })
})

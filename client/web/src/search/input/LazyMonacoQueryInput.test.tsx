import { createMemoryHistory } from 'history'
import { noop } from 'lodash'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'

import { SearchPatternType } from '../../graphql-operations'

import { PlainQueryInput } from './LazyMonacoQueryInput'

describe('PlainQueryInput', () => {
    const history = createMemoryHistory()
    test('empty', () =>
        expect(
            renderer
                .create(
                    <PlainQueryInput
                        history={history}
                        location={history.location}
                        queryState={{
                            query: '',
                        }}
                        patternType={SearchPatternType.regexp}
                        setPatternType={noop}
                        caseSensitive={true}
                        setCaseSensitivity={noop}
                        onChange={noop}
                        onSubmit={noop}
                        isLightTheme={false}
                        settingsCascade={{ subjects: [], final: {} }}
                        copyQueryButton={false}
                        showSearchContext={false}
                        showSearchContextManagement={false}
                        selectedSearchContextSpec=""
                        setSelectedSearchContextSpec={noop}
                        defaultSearchContextSpec=""
                        versionContext={undefined}
                        globbing={false}
                        enableSmartQuery={false}
                        fetchAutoDefinedSearchContexts={of([])}
                        fetchSearchContexts={() =>
                            of({
                                nodes: [],
                                pageInfo: {
                                    endCursor: null,
                                    hasNextPage: false,
                                },
                                totalCount: 0,
                            })
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot())

    test('with query', () =>
        expect(
            renderer
                .create(
                    <PlainQueryInput
                        history={history}
                        location={history.location}
                        queryState={{
                            query: 'repo:jsonrpc2 file:async.go asyncHandler',
                        }}
                        patternType={SearchPatternType.regexp}
                        setPatternType={noop}
                        caseSensitive={true}
                        setCaseSensitivity={noop}
                        onChange={noop}
                        onSubmit={noop}
                        isLightTheme={false}
                        settingsCascade={{ subjects: [], final: {} }}
                        copyQueryButton={false}
                        showSearchContext={false}
                        showSearchContextManagement={false}
                        selectedSearchContextSpec=""
                        setSelectedSearchContextSpec={noop}
                        defaultSearchContextSpec=""
                        versionContext={undefined}
                        globbing={false}
                        enableSmartQuery={false}
                        fetchAutoDefinedSearchContexts={of([])}
                        fetchSearchContexts={() =>
                            of({
                                nodes: [],
                                pageInfo: {
                                    endCursor: null,
                                    hasNextPage: false,
                                },
                                totalCount: 0,
                            })
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})

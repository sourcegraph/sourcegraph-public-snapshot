import { render } from '@testing-library/react'
import { createMemoryHistory, createLocation } from 'history'
import { noop } from 'lodash'
import React from 'react'
import { NEVER } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchPatternType } from '../../graphql-operations'

import { SearchResultsInfoBar, SearchResultsInfoBarProps } from './SearchResultsInfoBar'

const COMMON_PROPS: Omit<SearchResultsInfoBarProps, 'enableCodeMonitoring'> = {
    extensionsController,
    platformContext: { forceUpdateTooltip: noop, settings: NEVER },
    history: createMemoryHistory(),
    location: createLocation('/search'),
    authenticatedUser: { id: 'userID' },
    resultsFound: true,
    allExpanded: true,
    onExpandAllResultsToggle: noop,
    onSaveQueryClick: noop,
    stats: <div />,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    patternType: SearchPatternType.literal,
    caseSensitive: false,
}

describe('SearchResultsInfoBar', () => {
    test('code monitoring feature flag disabled', () => {
        expect(
            render(
                <SearchResultsInfoBar {...COMMON_PROPS} enableCodeMonitoring={false} query="foo type:diff" />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, cannot create monitor from query', () => {
        expect(
            render(<SearchResultsInfoBar {...COMMON_PROPS} enableCodeMonitoring={true} query="foo" />).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, can create monitor from query', () => {
        expect(
            render(
                <SearchResultsInfoBar {...COMMON_PROPS} enableCodeMonitoring={true} query="foo type:diff" />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, can create monitor from query, user not logged in', () => {
        expect(
            render(
                <SearchResultsInfoBar
                    {...COMMON_PROPS}
                    enableCodeMonitoring={true}
                    query="foo type:diff"
                    authenticatedUser={null}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('unauthenticated user', () => {
        expect(
            render(
                <SearchResultsInfoBar
                    {...COMMON_PROPS}
                    enableCodeMonitoring={true}
                    query="foo type:diff"
                    authenticatedUser={null}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})

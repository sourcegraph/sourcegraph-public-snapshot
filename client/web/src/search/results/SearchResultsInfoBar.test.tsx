import { createMemoryHistory, createLocation } from 'history'
import { noop } from 'lodash'
import { NEVER } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchPatternType } from '../../graphql-operations'

import { SearchResultsInfoBar, SearchResultsInfoBarProps } from './SearchResultsInfoBar'

const history = createMemoryHistory()
const COMMON_PROPS: Omit<SearchResultsInfoBarProps, 'enableCodeMonitoring'> = {
    extensionsController,
    platformContext: { forceUpdateTooltip: noop, settings: NEVER },
    history,
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
            renderWithBrandedContext(
                <SearchResultsInfoBar {...COMMON_PROPS} enableCodeMonitoring={false} query="foo type:diff" />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, cannot create monitor from query', () => {
        expect(
            renderWithBrandedContext(
                <SearchResultsInfoBar {...COMMON_PROPS} enableCodeMonitoring={true} query="foo" />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, can create monitor from query', () => {
        expect(
            renderWithBrandedContext(
                <SearchResultsInfoBar {...COMMON_PROPS} enableCodeMonitoring={true} query="foo type:diff" />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, can create monitor from query, user not logged in', () => {
        expect(
            renderWithBrandedContext(
                <SearchResultsInfoBar
                    {...COMMON_PROPS}
                    enableCodeMonitoring={true}
                    query="foo type:diff"
                    authenticatedUser={null}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('unauthenticated user', () => {
        expect(
            renderWithBrandedContext(
                <SearchResultsInfoBar
                    {...COMMON_PROPS}
                    enableCodeMonitoring={true}
                    query="foo type:diff"
                    authenticatedUser={null}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })
})

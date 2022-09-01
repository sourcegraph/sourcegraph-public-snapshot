import { createMemoryHistory, createLocation } from 'history'
import { noop } from 'lodash'
import { NEVER } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchPatternType } from '../../graphql-operations'

import { SearchResultsInfoBar, SearchResultsInfoBarProps } from './SearchResultsInfoBar'

const history = createMemoryHistory()
const COMMON_PROPS: Omit<SearchResultsInfoBarProps, 'enableCodeMonitoring'> = {
    extensionsController,
    platformContext: { settings: NEVER, sourcegraphURL: 'https://sourcegraph.com' },
    history,
    location: createLocation('/search'),
    authenticatedUser: { id: 'userID' },
    resultsFound: true,
    allExpanded: true,
    onExpandAllResultsToggle: noop,
    onSaveQueryClick: noop,
    stats: <div />,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    patternType: SearchPatternType.standard,
    caseSensitive: false,
}

const renderSearchResultsInfoBar = (
    props: Pick<SearchResultsInfoBarProps, 'enableCodeMonitoring'> & Partial<SearchResultsInfoBarProps>
) =>
    renderWithBrandedContext(
        <MockedTestProvider>
            <SearchResultsInfoBar {...COMMON_PROPS} {...props} />
        </MockedTestProvider>,
        { history }
    )

describe('SearchResultsInfoBar', () => {
    beforeAll(() => {
        window.context = {
            enableLegacyExtensions: true,
        } as any
    })

    test('code monitoring feature flag disabled', () => {
        expect(
            renderSearchResultsInfoBar({ query: 'foo type:diff', enableCodeMonitoring: false }).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, cannot create monitor from query', () => {
        expect(renderSearchResultsInfoBar({ query: 'foo', enableCodeMonitoring: true }).asFragment()).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, can create monitor from query', () => {
        expect(
            renderSearchResultsInfoBar({ query: 'foo type:diff', enableCodeMonitoring: true }).asFragment()
        ).toMatchSnapshot()
    })

    test('code monitoring feature flag enabled, can create monitor from query, user not logged in', () => {
        expect(
            renderSearchResultsInfoBar({
                query: 'foo type:diff',
                enableCodeMonitoring: true,
                authenticatedUser: null,
            }).asFragment()
        ).toMatchSnapshot()
    })

    test('unauthenticated user', () => {
        expect(
            renderSearchResultsInfoBar({
                query: 'foo type:diff',
                enableCodeMonitoring: true,
                authenticatedUser: null,
            }).asFragment()
        ).toMatchSnapshot()
    })
})

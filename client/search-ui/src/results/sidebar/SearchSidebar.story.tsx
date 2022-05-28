import { storiesOf } from '@storybook/react'
// We need to import `create` to make a mock store just for this story.
// eslint-disable-next-line no-restricted-imports
import create from 'zustand'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import {
    BuildSearchQueryURLParameters,
    InitialParametersSource,
    SearchPatternType,
    SearchQueryState,
    SearchQueryStateStoreProvider,
} from '@sourcegraph/search'
import { QuickLink, SearchScope } from '@sourcegraph/shared/src/schema/settings.schema'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { SearchSidebar, SearchSidebarProps } from './SearchSidebar'

const { add } = storiesOf('search-ui/results/sidebar/SearchSidebar', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/?node-id=1018%3A13883',
    },
    chromatic: { viewports: [544, 577, 993], disableSnapshot: false },
})

const mockUseQueryState = create<SearchQueryState>((set, get) => ({
    parametersSource: InitialParametersSource.DEFAULT,
    queryState: { query: '' },
    searchCaseSensitivity: false,
    searchPatternType: SearchPatternType.literal,
    searchQueryFromURL: '',
    setQueryState: queryStateUpdate => {
        if (typeof queryStateUpdate === 'function') {
            set({ queryState: queryStateUpdate(get().queryState) })
        } else {
            set({ queryState: queryStateUpdate })
        }
    },
    submitSearch: () => {},
}))

const defaultProps: SearchSidebarProps = {
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    selectedSearchContextSpec: 'global',
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    buildSearchURLQueryFromQueryState: (parameters: BuildSearchQueryURLParameters) => {
        const currentState = mockUseQueryState.getState()

        return buildSearchURLQuery(
            parameters.query,
            parameters.patternType ?? currentState.searchPatternType,
            parameters.caseSensitive ?? currentState.searchCaseSensitivity,
            parameters.searchContextSpec,
            parameters.searchParametersList
        )
    },
}

const quicklinks: QuickLink[] = [
    { name: 'Home', url: '/' },
    {
        name: 'This is a quicklink with a very long name lorem ipsum dolor sit amet',
        url: 'http://example.com',
        description: 'Example QuickLink',
    },
]

const scopes: SearchScope[] = [
    { name: 'This is a search scope with a very long name lorem ipsum dolor sit amet', value: 'repo:sourcegraph' },
    { name: 'All results', value: 'count:all' },
]

const filters: Filter[] = [
    {
        label: 'github.com/test/this-is-a-very-long-repo-name',
        value: 'repo:^github\\.com/test/this-is-a-very-long-repo-name$',
        count: 5,
        limitHit: false,
        kind: 'repo',
    },
    {
        label: 'gitlab.com/sourcegraph/sourcegraph',
        value: 'repo:^gitlab\\.com/sourcegraph/sourcegraph$',
        count: 201,
        limitHit: true,
        kind: 'repo',
    },
    {
        label: 'github.com/microsoft/vscode',
        value: 'repo:^github\\.com/microsoft/vscode$',
        count: 10,
        limitHit: true,
        kind: 'repo',
    },
    {
        label: 'bitbucket.com/test/test',
        value: 'repo:^bitbucket\\.com/test/test$',
        count: 1,
        limitHit: true,
        kind: 'repo',
    },
    {
        label: 'gitlab.sgdev.org/example/test',
        value: 'repo:^gitlab\\.sgdev\\.org/example/test$',
        count: 10,
        limitHit: true,
        kind: 'repo',
    },

    {
        label: 'lang:go',
        value: 'lang:go',
        count: 500,
        limitHit: true,
        kind: 'lang',
    },

    {
        label: 'lang:verylonglanguagenameloremipsumdolor',
        value: 'lang:verylonglanguagenameloremipsumdolor',
        count: 241,
        limitHit: false,
        kind: 'lang',
    },

    {
        label: '-file:_test\\.go$',
        value: '-file:_test\\.go$',
        count: 1230,
        limitHit: false,
        kind: 'file',
    },

    ...['typescript', 'javascript', 'c++', 'c', 'c#', 'python', 'ruby', 'haskell', 'java'].map(lang => ({
        label: `lang:${lang}`,
        value: `lang:${lang}`,
        count: 10,
        limitHit: true,
        kind: 'lang',
    })),
]

add('empty sidebar', () => (
    <BrandedStory>
        {() => (
            <SearchQueryStateStoreProvider useSearchQueryState={mockUseQueryState}>
                <SearchSidebar {...defaultProps} />
            </SearchQueryStateStoreProvider>
        )}
    </BrandedStory>
))

add('with everything', () => (
    <BrandedStory>
        {() => (
            <SearchQueryStateStoreProvider useSearchQueryState={mockUseQueryState}>
                <SearchSidebar
                    {...defaultProps}
                    settingsCascade={{ subjects: [], final: { quicklinks, 'search.scopes': scopes } }}
                    filters={filters}
                />
            </SearchQueryStateStoreProvider>
        )}
    </BrandedStory>
))

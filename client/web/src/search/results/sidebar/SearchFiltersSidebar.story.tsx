import type { Meta, StoryFn } from '@storybook/react'

import type { QuickLink, SearchScope } from '@sourcegraph/shared/src/schema/settings.schema'
import type { Filter } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../components/WebStory'
import { SearchPatternType } from '../../../graphql-operations'

import { SearchFiltersSidebar, type SearchFiltersSidebarProps } from './SearchFiltersSidebar'

const config: Meta = {
    title: 'web/search/results/sidebar/SearchFiltersSidebar',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/?node-id=1018%3A13883',
        },
        chromatic: { viewports: [544, 577, 993], disableSnapshot: false },
    },
}

export default config

const defaultProps: SearchFiltersSidebarProps = {
    liveQuery: '',
    submittedURLQuery: '',
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    onNavbarQueryChange: () => {},
    onSearchSubmit: () => {},
    selectedSearchContextSpec: 'global',
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    setSidebarCollapsed: () => {},
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
        label: 'bitbucket.org/com/test',
        value: 'repo:^bitbucket\\.org/com/test$',
        count: 1,
        limitHit: true,
        kind: 'repo',
    },
    {
        label: 'bitbucket.org/org/test',
        value: 'repo:^bitbucket\\.org/org/test$',
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

    ...['TypeScript', 'JavaScript', 'C++', 'C', 'C#', 'Python', 'Ruby', 'Haskell', 'Java'].map(lang => ({
        label: lang,
        value: `lang:${lang.toLowerCase()}`,
        count: 10,
        limitHit: true,
        kind: 'lang' as Filter['kind'],
    })),
]

export const EmptySidebar: StoryFn = () => <WebStory>{() => <SearchFiltersSidebar {...defaultProps} />}</WebStory>

EmptySidebar.storyName = 'empty sidebar'

export const WithEverything: StoryFn = () => (
    <WebStory>
        {() => (
            <SearchFiltersSidebar
                {...defaultProps}
                settingsCascade={{ subjects: [], final: { quicklinks, 'search.scopes': scopes } }}
                filters={filters}
            />
        )}
    </WebStory>
)

WithEverything.storyName = 'with everything'

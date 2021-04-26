import { storiesOf } from '@storybook/react'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { SearchPatternType } from '../../../../graphql-operations'
import { QuickLink, SearchScope } from '../../../../schema/settings.schema'
import { Filter } from '../../../stream'

import { SearchSidebar, SearchSidebarProps } from './SearchSidebar'

const { add } = storiesOf('web/search/results/streaming/sidebar/SearchSidebar', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/?node-id=1018%3A13883',
    },
})

const defaultProps: SearchSidebarProps = {
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    versionContext: undefined,
    selectedSearchContextSpec: 'global',
    query: '',
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    telemetryService: NOOP_TELEMETRY_SERVICE,
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
        label: 'sourcegraph/sourcegraph',
        value: 'repo:^sourcegraph/sourcegraph$',
        count: 201,
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
]

add('empty sidebar', () => <WebStory>{() => <SearchSidebar {...defaultProps} />}</WebStory>)

add('with everything', () => (
    <WebStory>
        {() => (
            <SearchSidebar
                {...defaultProps}
                settingsCascade={{ subjects: [], final: { quicklinks, 'search.scopes': scopes } }}
                filters={filters}
            />
        )}
    </WebStory>
))

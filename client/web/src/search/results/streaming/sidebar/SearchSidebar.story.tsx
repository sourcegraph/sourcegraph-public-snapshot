import { storiesOf } from '@storybook/react'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { SearchPatternType } from '../../../../graphql-operations'
import { QuickLink, SearchScope } from '../../../../schema/settings.schema'

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
    { name: 'Example', url: 'http://example.com', description: 'Example QuickLink' },
]

const scopes: SearchScope[] = [
    { name: 'Sourcegraph repos', value: 'repo:sourcegraph' },
    { name: 'All results', value: 'count:all' },
]

add('empty sidebar', () => <WebStory>{() => <SearchSidebar {...defaultProps} />}</WebStory>)

add('with quicklinks', () => (
    <WebStory>
        {() => <SearchSidebar {...defaultProps} settingsCascade={{ subjects: [], final: { quicklinks } }} />}
    </WebStory>
))

add('with scopes', () => (
    <WebStory>
        {() => (
            <SearchSidebar {...defaultProps} settingsCascade={{ subjects: [], final: { 'search.scopes': scopes } }} />
        )}
    </WebStory>
))

add('with everything', () => (
    <WebStory>
        {() => (
            <SearchSidebar
                {...defaultProps}
                settingsCascade={{ subjects: [], final: { quicklinks, 'search.scopes': scopes } }}
            />
        )}
    </WebStory>
))

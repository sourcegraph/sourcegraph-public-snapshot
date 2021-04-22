import { storiesOf } from '@storybook/react'
import React from 'react'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../../components/WebStory'
import { SearchPatternType } from '../../../../graphql-operations'
import { QuickLink } from '../../../../schema/settings.schema'

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
}

const quicklinks: QuickLink[] = [
    { name: 'Home', url: '/' },
    { name: 'Example', url: 'http://example.com', description: 'Example QuickLink' },
]

add('empty sidebar', () => <WebStory>{() => <SearchSidebar {...defaultProps} />}</WebStory>)

add('with quicklinks', () => (
    <WebStory>
        {() => <SearchSidebar {...defaultProps} settingsCascade={{ subjects: [], final: { quicklinks } }} />}
    </WebStory>
))

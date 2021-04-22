import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'
import { SearchPatternType } from '../../../../graphql-operations'

import { SearchSidebar, SearchSidebarProps } from './SearchSidebar'

const { add } = storiesOf('web/search/results/streaming/SearchSidebar', module).addParameters({
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
}

add('empty sidebar', () => <WebStory>{() => <SearchSidebar {...defaultProps} />}</WebStory>)

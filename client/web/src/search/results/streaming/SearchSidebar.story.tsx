import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { SearchSidebar } from './SearchSidebar'

const { add } = storiesOf('web/search/results/streaming/SearchSidebar', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/?node-id=1018%3A13883',
    },
})

add('empty sidebar', () => <WebStory>{() => <SearchSidebar />}</WebStory>)

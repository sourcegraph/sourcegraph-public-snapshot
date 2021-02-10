import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { SearchContextMenu } from './SearchContextMenu'

const { add } = storiesOf('web/search/input/SearchContextMenu', module)
    .addParameters({
        chromatic: { viewports: [500] },
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/4Fy9rURbfF2bsl4BvYunUO/RFC-261-Search-Contexts?node-id=581%3A4754',
        },
    })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ position: 'static' }}>
            {story()}
        </div>
    ))

add('default', () => <WebStory>{() => <SearchContextMenu />}</WebStory>, {})

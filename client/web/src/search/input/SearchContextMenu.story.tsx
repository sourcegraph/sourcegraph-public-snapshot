import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { SearchContextMenu } from './SearchContextMenu'

const { add } = storiesOf('web/search/input/SearchContextMenu', module)
    .addParameters({ chromatic: { viewports: [400] } })
    .addDecorator(story => (
        <div className="dropdown-menu show" style={{ width: '400px', position: 'static' }}>
            {story()}
        </div>
    ))

add('default', () => <WebStory>{() => <SearchContextMenu />}</WebStory>, {})

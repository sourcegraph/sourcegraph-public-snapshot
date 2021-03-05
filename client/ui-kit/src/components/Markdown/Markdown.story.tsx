import React from 'react'
import { createMemoryHistory } from 'history'
import { addStory } from '../../storybook-helpers'
import { Markdown } from './Markdown'

addStory('web', 'Markdown').add('Basic', () => {
    const history = createMemoryHistory()

    return (
        <div style={{ background: 'white', padding: '20px ' }}>
            <Markdown history={history} dangerousInnerHTML="Hello world!" />
        </div>
    )
})

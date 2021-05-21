import { storiesOf } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Container } from './Container'

const { add } = storiesOf('wildcard/Container', module).addDecorator(story => (
    <BrandedStory styles={webStyles}>{() => <div className="container web-content mt-3">{story()}</div>}</BrandedStory>
))

add(
    'Multiple containers',
    () => (
        <div className="row">
            <div className="col-6">
                <Container>
                    <h4>Watch for AWS secrets in commits</h4>
                    <p>
                        Use a search query to watch for new search results, and choose how to receive notifications in
                        response.
                    </p>
                    <button className="btn btn-sm btn-secondary">View in docs →</button>
                </Container>
            </div>
            <div className="col-6">
                <Container>
                    <h4>Watch for AWS secrets in commits</h4>
                    <p>
                        Use a search query to watch for new search results, and choose how to receive notifications in
                        response.
                    </p>
                    <button className="btn btn-sm btn-secondary">View in docs →</button>
                </Container>
            </div>
        </div>
    ),
    {
        design: [
            {
                type: 'figma',
                name: 'Figma Redesign',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/?node-id=1478%3A3044',
            },
        ],
    }
)

import { storiesOf } from '@storybook/react'

import { Card } from '@sourcegraph/wildcard'

import { WebStory } from '../components/WebStory'

import { OrDivider } from './OrDivider'

const { add } = storiesOf('web/OrDivider', module).addDecorator(story => <div className="p-3 container">{story()}</div>)

add('Alone', () => (
    <WebStory>
        {() => (
            <Card className="border-0">
                <OrDivider />
            </Card>
        )}
    </WebStory>
))

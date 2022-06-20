import { storiesOf } from '@storybook/react'

import { Text } from '@sourcegraph/wildcard'

import { LoaderButton } from './LoaderButton'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/LoaderButton', module).addDecorator(story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
))

add('Inline', () => (
    <WebStory>
        {() => (
            <Text>
                <LoaderButton loading={true} label="loader button" variant="primary" />
            </Text>
        )}
    </WebStory>
))

add('Block', () => (
    <WebStory>
        {() => <LoaderButton loading={true} label="loader button" className="btn-block" variant="primary" />}
    </WebStory>
))

add('With label', () => (
    <WebStory>
        {() => (
            <LoaderButton
                alwaysShowLabel={true}
                loading={true}
                label="loader button"
                className="btn-block"
                variant="primary"
            />
        )}
    </WebStory>
))

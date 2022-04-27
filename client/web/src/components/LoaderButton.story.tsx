import { storiesOf } from '@storybook/react'

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
            <p>
                <LoaderButton loading={true} label="loader button" variant="primary" />
            </p>
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

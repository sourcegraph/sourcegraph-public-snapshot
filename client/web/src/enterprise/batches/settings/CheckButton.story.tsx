import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { CheckButton } from './CheckButton'

const { add } = storiesOf('web/batches/settings/CheckButton', module)

add('Initial', () => (
    <WebStory>
        {props => (
            <CheckButton {...props} label="Checks the state of something" onClick={action('onClick')} loading={false} />
        )}
    </WebStory>
))

add('Checking', () => (
    <WebStory>
        {props => <CheckButton {...props} label="Checks the state of something" onClick={() => {}} loading={true} />}
    </WebStory>
))

add('Success', () => (
    <WebStory>
        {props => (
            <CheckButton
                {...props}
                label="Checks the state of something"
                onClick={() => {}}
                loading={false}
                successMessage="Credential is valid"
            />
        )}
    </WebStory>
))

add('Failed', () => (
    <WebStory>
        {props => (
            <CheckButton
                {...props}
                label="Checks the state of something"
                onClick={() => {}}
                loading={false}
                failedMessage="The credential is not valid. Something went wrong when connecting to the code host"
            />
        )}
    </WebStory>
))

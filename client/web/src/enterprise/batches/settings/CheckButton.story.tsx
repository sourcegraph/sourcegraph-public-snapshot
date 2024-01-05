import { action } from '@storybook/addon-actions'
import type { Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { CheckButton } from './CheckButton'

const config: Meta = {
    title: 'web/batches/settings/CheckButton',
}

export default config

export const Initial: StoryFn = () => (
    <WebStory>
        {props => (
            <CheckButton {...props} label="Checks the state of something" onClick={action('onClick')} loading={false} />
        )}
    </WebStory>
)

export const Checking: StoryFn = () => (
    <WebStory>
        {props => <CheckButton {...props} label="Checks the state of something" onClick={() => {}} loading={true} />}
    </WebStory>
)

export const Success: StoryFn = () => (
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
)

export const Failed: StoryFn = () => (
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
)

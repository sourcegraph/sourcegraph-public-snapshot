import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { IDEExtensionAlert } from './IdeExtensionAlert'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/repo/actions',
    decorators: [decorator],
    parameters: {
        component: IDEExtensionAlert,
    },
}

export default config

export const IDEExtensionAlertDefault: Story = () => (
    <IDEExtensionAlert page="search" onAlertDismissed={action('onAlertDismissed')} />
)
IDEExtensionAlertDefault.storyName = 'IDEExtensionAlert'

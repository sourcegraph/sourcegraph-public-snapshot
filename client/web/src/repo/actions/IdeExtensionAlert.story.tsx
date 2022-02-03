import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

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

export const Default: Story = () => <IDEExtensionAlert className="" onAlertDismissed={action('onAlertDismissed')} />
Default.storyName = 'IDEExtensionAlert'

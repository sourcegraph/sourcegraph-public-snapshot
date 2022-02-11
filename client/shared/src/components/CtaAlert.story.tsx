import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

import { ExtensionRadialGradientIcon } from '@sourcegraph/web/src/components/CtaIcons'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { CtaAlert } from './CtaAlert'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'shared',
    decorators: [decorator],
    parameters: {
        component: CtaAlert,
        chromatic: {
            enableDarkMode: true,
        },
    },
}

export default config

export const Default: Story = () => (
    <CtaAlert
        title="Title"
        description="Description"
        cta={{ label: 'Label', href: '#' }}
        icon={<ExtensionRadialGradientIcon />}
        onClose={() => {}}
    />
)
Default.storyName = 'CtaAlert'

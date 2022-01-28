import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'
import { ExtensionRadialGradientIcon } from '@sourcegraph/web/src/search/CtaIcons'

import { CtaAlert } from './CtaAlert'

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'shared',
    decorators: [decorator],
    parameters: {
        component: CtaAlert,
    },
}

export default config

export const Default: Story = () => (
    <CtaAlert
        title="Title"
        description="Description"
        cta={{ label: 'Label', href: '#' }}
        icon={<ExtensionRadialGradientIcon />}
        className=""
        onClose={() => {}}
    />
)
Default.storyName = 'CtaAlert'

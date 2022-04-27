import { DecoratorFn, Meta, Story } from '@storybook/react'

// eslint-disable-next-line no-restricted-imports
import { ExtensionRadialGradientIcon } from '@sourcegraph/web/src/components/CtaIcons'
// eslint-disable-next-line no-restricted-imports
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
            disableSnapshot: false,
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

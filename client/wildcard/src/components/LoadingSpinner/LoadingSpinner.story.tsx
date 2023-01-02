import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '../../stories/BrandedStory'

import { LoadingSpinner } from './LoadingSpinner'

const config: Meta = {
    title: 'wildcard/LoadingSpinner',
    component: LoadingSpinner,

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: LoadingSpinner,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
    argTypes: {
        inline: {
            control: { type: 'boolean' },
            defaultValue: true,
        },
    },
}

export default config

export const Simple: Story = (args = {}) => <LoadingSpinner inline={args.inline} />

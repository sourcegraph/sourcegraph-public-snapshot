import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { LoadingSpinner } from './LoadingSpinner'

const config: Meta = {
    title: 'wildcard/LoadingSpinner',
    component: LoadingSpinner,

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

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

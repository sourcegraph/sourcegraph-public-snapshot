import { boolean } from '@storybook/addon-knobs'
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
}

export default config

export const Simple: Story = () => <LoadingSpinner inline={boolean('inline', true)} />

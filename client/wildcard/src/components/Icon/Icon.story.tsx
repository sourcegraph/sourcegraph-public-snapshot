import { Story, Meta } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { SourcegraphIcon } from '../SourcegraphIcon'

import { Icon } from './Icon'

const config: Meta = {
    title: 'wildcard/Icon',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Icon,
    },
}
export default config

export const Simple: Story = () => (
    <>
        <h3>Small Icon</h3>
        <Icon svg={<SourcegraphIcon />} size="sm" />

        <h3>Medium Icon</h3>
        <Icon svg={<SourcegraphIcon />} size="md" />
    </>
)

import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { PRODUCT_STATUSES } from '@sourcegraph/wildcard/src/components/Badge/constants'

import { FeedbackBadge } from '.'

const config: Meta = {
    title: 'wildcard/FeedbackBadge',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        component: FeedbackBadge,
    },
}

export default config

export const FeedbackBadgeExample: Story = () => (
    <>
        <h1>FeedbackBadges</h1>
        <p>Our badges come in different status.</p>
        {PRODUCT_STATUSES.map(status => (
            <FeedbackBadge
                className="mb-2"
                status={status}
                key={status}
                feedback={{ mailto: 'support@sourcegraph.com' }}
            />
        ))}
    </>
)

import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { PRODUCT_STATUSES } from '@sourcegraph/wildcard'

import { Typography } from '../..'

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
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A6149',
            },

            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A6447',
            },
        ],
    },
}

export default config

export const FeedbackBadgeExample: Story = () => (
    <>
        <Typography.H1>FeedbackBadges</Typography.H1>
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

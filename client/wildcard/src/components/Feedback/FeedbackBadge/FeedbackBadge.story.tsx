import type { Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '../../../stories/BrandedStory'
import { PRODUCT_STATUSES } from '../../Badge'
import { H1, Text } from '../../Typography'

import { FeedbackBadge } from '.'

const config: Meta = {
    title: 'wildcard/FeedbackBadge',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
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

export const FeedbackBadgeExample: StoryFn = () => (
    <>
        <H1>FeedbackBadges</H1>
        <Text>Our badges come in different status.</Text>
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

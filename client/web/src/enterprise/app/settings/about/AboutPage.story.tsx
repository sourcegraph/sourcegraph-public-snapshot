import type { Meta, StoryObj } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { AboutTab, type AboutTabProps } from './AboutPage'

const meta: Meta<AboutTabProps> = {
    component: AboutTab,
    args: {
        version: '1.2.3',
    },
    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],
}

export default meta

type Story = StoryObj<typeof AboutTab>

export const AboutPage: Story = {}

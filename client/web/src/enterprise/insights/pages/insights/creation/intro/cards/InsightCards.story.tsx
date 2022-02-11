import { Meta, Story } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../../../../components/WebStory'

import { CaptureGroupInsightCard, LangStatsInsightCard, SearchInsightCard } from './InsightCards'

export default {
    title: 'web/insights/InsightCards',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            disableSnapshot: false,
        },
    },
} as Meta

export const InsightCards: Story = () => (
    <section className="row">
        <article className="col-sm-4">
            <h2>Search Insight Card</h2>
            <SearchInsightCard />
        </article>
        <article className="col-sm-4">
            <h2>Language Stats Insight Card</h2>
            <LangStatsInsightCard />
        </article>
        <article className="col-sm-4">
            <h2>Capture Group Insight Card</h2>
            <CaptureGroupInsightCard />
        </article>
    </section>
)

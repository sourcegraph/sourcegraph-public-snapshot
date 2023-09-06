import type { Meta, Story } from '@storybook/react'

import { H2 } from '@sourcegraph/wildcard'

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
            <H2>Search Insight Card</H2>
            <SearchInsightCard />
        </article>
        <article className="col-sm-4">
            <H2>Language Stats Insight Card</H2>
            <LangStatsInsightCard />
        </article>
        <article className="col-sm-4">
            <H2>Capture Group Insight Card</H2>
            <CaptureGroupInsightCard />
        </article>
    </section>
)

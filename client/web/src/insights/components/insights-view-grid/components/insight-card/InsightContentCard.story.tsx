import { storiesOf } from '@storybook/react'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'

import { InsightErrorContent } from './components/insight-error-content/InsightErrorContent'
import { InsightLoadingContent } from './components/insight-loading-content/InsightLoadingContent'
import { InsightContentCard } from './InsightContentCard'

const { add } = storiesOf('web/insights/InsightContentCard', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Loading insight', () => (
    <InsightContentCard
        style={{ width: '400px', height: '400px' }}
        insight={{ id: 'searchInsights.insight.id', view: undefined }}
        onDelete={() => Promise.resolve()}
        hasContextMenu={true}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    >
        <InsightLoadingContent text="Loading insight" subTitle="searchInsights.insight.id" icon={PuzzleIcon} />
    </InsightContentCard>
))

add('Errored insight', () => (
    <InsightContentCard
        style={{ width: '400px', height: '400px' }}
        insight={{
            id: 'searchInsights.insight.id',
            view: new Error("BE couldn't load this Insight"),
        }}
        hasContextMenu={true}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    >
        <InsightErrorContent
            title="searchInsights.insight.id"
            error={new Error("We couldn't find code insight")}
            icon={PuzzleIcon}
        />
    </InsightContentCard>
))

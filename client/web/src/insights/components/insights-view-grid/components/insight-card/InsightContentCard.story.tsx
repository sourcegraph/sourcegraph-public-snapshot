import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { ViewInsightProviderSourceType } from '../../../../core/backend/types'

import { InsightContentCard } from './InsightContentCard'

const { add } = storiesOf('web/insights/InsightContentCard', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Loading insight', () => (
    <InsightContentCard
        style={{ width: '400px', height: '400px' }}
        insight={{ id: 'searchInsights.insight.id', source: ViewInsightProviderSourceType.Extension, view: undefined }}
        onDelete={() => Promise.resolve()}
        hasContextMenu={true}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    />
))

add('Errored insight', () => (
    <InsightContentCard
        style={{ width: '400px', height: '400px' }}
        insight={{
            id: 'searchInsights.insight.id',
            source: ViewInsightProviderSourceType.Extension,
            view: new Error("BE couldn't load this Insight"),
        }}
        hasContextMenu={true}
        onDelete={() => Promise.resolve()}
        telemetryService={NOOP_TELEMETRY_SERVICE}
    />
))

import { storiesOf } from '@storybook/react'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useState } from 'react'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK } from '../../../mocks/charts-content'
import { ViewContent } from '../../content/view-content/ViewContent'
import { ViewErrorContent } from '../../content/view-error-content/ViewErrorContent'
import { ViewLoadingContent } from '../../content/view-loading-content/ViewLoadingContent'

import { ViewCard } from './ViewCard'

const { add } = storiesOf('web/insights/InsightContentCard', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Loading insight', () => (
    <ViewCard
        style={{ width: '400px', height: '400px' }}
        insight={{ id: 'searchInsights.insight.id', view: undefined }}
    >
        <ViewLoadingContent text="Loading insight" subTitle="searchInsights.insight.id" icon={PuzzleIcon} />
    </ViewCard>
))

add('Errored insight', () => (
    <ViewCard
        style={{ width: '400px', height: '400px' }}
        insight={{
            id: 'searchInsights.insight.id',
            view: new Error("BE couldn't load this Insight"),
        }}
    >
        <ViewErrorContent
            title="searchInsights.insight.id"
            error={new Error("We couldn't find code insight")}
            icon={PuzzleIcon}
        />
    </ViewCard>
))

const EMPTY_INSIGHT_VIEW: ViewProviderResult = {
    id: 'searchInsights.insight.id',
    view: {
        title: 'Insight title',
        content: [LINE_CHART_CONTENT_MOCK],
    },
}

add('Content insight', () => (
    <ViewCard
        style={{ width: '363px', height: '312px' }}
        contextMenu={<DotsVerticalIcon size={16} />}
        actions={<FilterOutlineIcon className="mr-1" size="1rem" />}
        insight={EMPTY_INSIGHT_VIEW}
    >
        <ViewContent
            telemetryService={NOOP_TELEMETRY_SERVICE}
            viewContent={[LINE_CHART_CONTENT_MOCK]}
            viewID={EMPTY_INSIGHT_VIEW.id}
        />
    </ViewCard>
))

const VIEW_CONTENT = [LINE_CHART_CONTENT_MOCK]
const TWO_LINES_TITLE_INSIGHT_VIEW: ViewProviderResult = {
    id: 'searchInsights.insight.id',
    view: {
        title: 'Insight title',
        content: VIEW_CONTENT,
    },
}

add('Content insight with re-fetching state', () => {
    const [refetching, setRefetching] = useState(false)

    return (
        <>
            <ViewCard
                style={{ width: '363px', height: '312px' }}
                contextMenu={<DotsVerticalIcon size={16} />}
                actions={<FilterOutlineIcon className="mr-1" size="1rem" />}
                insight={TWO_LINES_TITLE_INSIGHT_VIEW}
                reFetchingStatus={refetching}
            >
                <ViewContent
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    viewContent={VIEW_CONTENT}
                    viewID={EMPTY_INSIGHT_VIEW.id}
                />
            </ViewCard>

            <label className="mt-3 d-flex align-items-center">
                <span className="mr-2">Refetching</span>
                <input type="checkbox" checked={refetching} onChange={event => setRefetching(event.target.checked)} />
            </label>
        </>
    )
})

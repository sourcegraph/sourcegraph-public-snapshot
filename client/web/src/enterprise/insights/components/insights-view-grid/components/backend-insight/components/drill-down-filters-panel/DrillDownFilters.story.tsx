import { useRef } from 'react'

import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../../../../components/WebStory'
import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../../graphql-operations'
import { BackendInsight, InsightExecutionType, InsightFilters, InsightType } from '../../../../../../core'
import { DrillDownFiltersPopover } from '../drill-down-filters-popover/DrillDownFiltersPopover'

const defaultStory: Meta = {
    title: 'DrillDownFilters',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

export const DrillDownPopover: Story = () => {
    const exampleReference = useRef(null)
    const initialFiltersValue: InsightFilters = {
        excludeRepoRegexp: 'EXCLUDE',
        includeRepoRegexp: '',
        context: '',
    }
    const insight: BackendInsight = {
        id: 'example',
        title: 'Example Insight',
        repositories: [],
        type: InsightType.CaptureGroup,
        executionType: InsightExecutionType.Backend,
        step: {},
        isFrozen: false,
        query: '',
        filters: initialFiltersValue,
        dashboardReferenceCount: 0,
        dashboards: [],
    }

    return (
        <DrillDownFiltersPopover
            isOpen={true}
            anchor={exampleReference}
            initialFiltersValue={initialFiltersValue}
            originalFiltersValue={initialFiltersValue}
            insight={insight}
            onFilterChange={log('onFilterChange')}
            onFilterSave={log('onFilterSave')}
            onInsightCreate={log('onInsightCreate')}
            onVisibilityChange={log('onVisibilityChange')}
            originalSeriesDisplayOptions={{
                limit: 20,
                sortOptions: {
                    direction: SeriesSortDirection.DESC,
                    mode: SeriesSortMode.RESULT_COUNT,
                },
            }}
            onSeriesDisplayOptionsChange={log('onSeriesDisplayOptionsChange')}
        />
    )
}

// eslint-disable-next-line arrow-body-style
const log = (methodName: string) => {
    return function (args: unknown) {
        // eslint-disable-next-line prefer-rest-params
        console.log(methodName, [...arguments])
    }
}

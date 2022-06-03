import { Meta, Story } from '@storybook/react'
import { useRef } from 'react'
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
        excludeRepoRegexp: '',
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
        filters: {
            excludeRepoRegexp: '',
            includeRepoRegexp: '',
            context: '',
        },
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

const log = (methodName: string) => {
    return function (args: unknown) {
        console.log(methodName, [...arguments])
    }
}

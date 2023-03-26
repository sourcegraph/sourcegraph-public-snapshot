import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { SearchAggregationResultProps } from './SearchAggregationResult'

export * from './types'
export * from './hooks'

export { AggregationChartCard, getAggregationData } from './components/aggregation-chart-card/AggregationChartCard'
export { AggregationModeControls } from './components/aggregation-mode-controls/AggregationModeControls'
export { AggregationLimitLabel } from './components/AggregationLimitLabel'
export { GroupResultsPing } from './pings'

export const SearchAggregationResult = !process.env.DISABLE_SEARCH_AGGREGATIONS
    ? lazyComponent<SearchAggregationResultProps, 'SearchAggregationResult'>(
          () => import('./SearchAggregationResult'),
          'SearchAggregationResult'
      )
    : null

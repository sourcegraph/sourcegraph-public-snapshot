import type { SearchBasedInsightSeries } from '../../../core'

export interface EditableDataSeries extends SearchBasedInsightSeries {
    valid: boolean
    edit: boolean
    autofocus: boolean
}

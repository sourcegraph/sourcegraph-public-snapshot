import { InsightVisibility } from '../../../../core/types'

export interface LangStatsCreationFormFields {
    title: string
    repository: string
    threshold: number
    visibility: InsightVisibility
}

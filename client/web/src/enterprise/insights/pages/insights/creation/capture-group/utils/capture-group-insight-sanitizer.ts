import { getSanitizedRepositories } from '../../../../../components/creation-ui-kit/sanitizers/repositories'
import { CaptureGroupInsight, InsightExecutionType, InsightType } from '../../../../../core/types'
import { CaptureGroupFormFields } from '../types'

export function getSanitizedCaptureGroupInsight(values: CaptureGroupFormFields): CaptureGroupInsight {
    return {
        title: values.title.trim(),
        query: values.groupSearchQuery.trim(),
        repositories: getSanitizedRepositories(values.repositories),
        viewType: InsightType.CaptureGroup,
        type: InsightExecutionType.Backend,
        id: '',
        visibility: '',
        step: { [values.step]: +values.stepValue },
    }
}

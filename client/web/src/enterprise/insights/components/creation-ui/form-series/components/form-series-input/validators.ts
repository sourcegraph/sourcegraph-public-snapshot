import { searchQueryValidator } from '../../../../../pages/insights/creation/capture-group/utils/search-query-validator'
import { createRequiredValidator, ValidationResult } from '../../../../form'

export const requiredNameField = createRequiredValidator('Name is a required field for data series.')

export const validQuery = (value: string | undefined, validity: ValidityState | null | undefined): ValidationResult => {
    const result = createRequiredValidator('Query is a required field for data series.')(value, validity)

    if (result) {
        return result
    }

    const { isNotContext, isNotRepo } = searchQueryValidator(value || '', true)

    if (!isNotContext) {
        return 'The `context:` filter is not supported; instead, run over all repositories and use the `context:` on the filter panel after creation'
    }

    if (!isNotRepo) {
        return 'Do not include a `repo:` filter; add targeted repositories above, or filter repos on the filter panel after creation'
    }
}

import { type Validator, composeValidators, createRequiredValidator } from '@sourcegraph/wildcard'

import { insightStepValueValidator, insightTitleValidator } from '../../../../../components'
import { type Checks, searchQueryValidator } from '../utils/search-query-validator'

const validateQueryCheck: Validator<string, Checks> = (value: string | undefined) => {
    const validatedChecks = searchQueryValidator(value)
    const allChecksPassed = Object.values(validatedChecks).every(Boolean)

    if (!allChecksPassed) {
        return { errorMessage: 'Query is not valid', context: validatedChecks }
    }

    return { context: validatedChecks }
}

export const QUERY_VALIDATORS = composeValidators<string, Checks>([
    createRequiredValidator('Query is a required field.'),
    validateQueryCheck,
])

export const TITLE_VALIDATORS = insightTitleValidator
export const STEP_VALIDATORS = insightStepValueValidator

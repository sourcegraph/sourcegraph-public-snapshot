import { renderError } from '@sourcegraph/branded/src/components/alerts'

import { Validator } from '../../../../../../components/form/hooks/useField'
import { AsyncValidator } from '../../../../../../components/form/hooks/utils/use-async-validation'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { fetchRepositories } from '../../../../../../core/backend/gql-backend/methods/get-built-in-insight-data/utils/fetch-repositories'

export const repositoriesFieldValidator: Validator<string> = value => {
    if (value !== undefined && value.trim() === '') {
        return 'Repositories is a required field for code insight.'
    }

    return
}

export const thresholdFieldValidator = createRequiredValidator('Threshold is a required field for code insight.')

// [TODO] [VK] Move this validator behind insight api context for better testing approach
export const repositoryFieldAsyncValidator: AsyncValidator<string> = async value => {
    if (value === undefined || value.trim() === '') {
        return
    }

    try {
        const repositories = await fetchRepositories([value.trim()]).toPromise()

        if (!repositories[0]?.name) {
            return `We couldn't find the repository ${value}. Please ensure the repository exists.`
        }

        return
    } catch (error) {
        return renderError(error)
    }
}

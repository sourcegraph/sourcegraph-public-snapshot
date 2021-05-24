import { renderError } from '../../../../../../components/alerts'
import { AsyncValidator } from '../../../../../components/form/hooks/utils/use-async-validation'
import { createRequiredValidator } from '../../../../../components/form/validators'
import { fetchRepositories } from '../../../../../core/backend/requests/fetch-repositories'

export const repositoriesFieldValidator = createRequiredValidator('Repositories is a required field for code insight.')
export const thresholdFieldValidator = createRequiredValidator('Threshold is a required field for code insight.')

// [TODO] [VK] Move this validator behind insight api context for better testing approach
export const repositoryFieldAsyncValidator: AsyncValidator<string> = async value => {
    if (!value) {
        return
    }

    try {
        const repositories = await fetchRepositories([value.trim()]).toPromise()

        if (!repositories[0]) {
            return `We couldn't find the repository ${value}. Please ensure the repository exists.`
        }

        return
    } catch (error) {
        return renderError(error)
    }
}

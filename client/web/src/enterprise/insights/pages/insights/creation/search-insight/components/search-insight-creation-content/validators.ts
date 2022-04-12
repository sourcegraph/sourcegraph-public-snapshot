import { renderError } from '@sourcegraph/branded/src/components/alerts'
import { dedupeWhitespace } from '@sourcegraph/common'

import { getSanitizedRepositories } from '../../../../../../components/creation-ui-kit'
import { Validator } from '../../../../../../components/form/hooks/useField'
import { AsyncValidator } from '../../../../../../components/form/hooks/utils/use-async-validation'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { fetchRepositories } from '../../../../../../core/backend/gql-backend/methods/get-built-in-insight-data/utils/fetch-repositories'
import { EditableDataSeries } from '../../types'

export const repositoriesFieldValidator: Validator<string> = value => {
    if (value !== undefined && dedupeWhitespace(value).trim() === '') {
        return 'Repositories is a required field.'
    }

    return
}

export const requiredStepValueField = createRequiredValidator('Please specify a step between points.')

/**
 * Custom validator for chart series. Since series has complex type
 * we can't validate this with standard validators.
 */
export const seriesRequired: Validator<EditableDataSeries[]> = series => {
    if (!series || series.length === 0) {
        return 'No series defined. You must add at least one series to create a code insight.'
    }

    if (series.some(series => !series.valid)) {
        return 'Some series is invalid. Remove or edit the invalid series.'
    }

    return
}

export const repositoriesExistValidator: AsyncValidator<string> = async value => {
    if (!value) {
        return
    }

    try {
        const repositoryNames = getSanitizedRepositories(value)

        if (repositoryNames.length === 0) {
            return
        }

        const repositories = await fetchRepositories(repositoryNames).toPromise()
        const nullRepositoryIndex = repositories.findIndex(
            (repo, index) => !repo || repo.name !== repositoryNames[index]
        )

        if (nullRepositoryIndex !== -1) {
            const repoName = repositoryNames[nullRepositoryIndex]

            return `We couldn't find the repository ${repoName}. Please ensure the repository exists.`
        }

        return
    } catch (error) {
        return renderError(error)
    }
}

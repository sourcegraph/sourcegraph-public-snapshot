import { renderError } from '@sourcegraph/branded/src/components/alerts'
import { dedupeWhitespace } from '@sourcegraph/common'

import { Validator, AsyncValidator } from '../../form'
import { createRequiredValidator } from '../../form/hooks/validators'
import { EditableDataSeries } from '../form-series'
import { getSanitizedRepositories } from '../sanitizers'

import { fetchRepositories } from './fetch-repositories'

// Group of shared Creation UI/Edit UI insight validators.

/**
 * Primarily used in any place where we edit or create insights, like creation ui page
 * or drill down insight creation flow.
 */
export const insightTitleValidator = createRequiredValidator('Title is a required field.')

/**
 * Primarily used in creation and edit insight pages and also on the landing page where
 * we have a creation UI insight sandbox demo widget.
 */
export const insightRepositoriesValidator: Validator<string> = value => {
    if (value !== undefined && dedupeWhitespace(value).trim() === '') {
        return 'Repositories is a required field.'
    }

    return
}

/**
 * Check that repositories exist on the backend. It takes a string: "repo1, repo2, ..."
 * and checks their existence on the backend. If one of the repositions doesn't exist it
 * will return error message.
 */
export const insightRepositoriesAsyncValidator: AsyncValidator<string> = async value => {
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

/**
 * Validator that should be used for the time interval settings control.
 * Like Search-Based or Capture group insights time interval controls.
 */
export const insightStepValueValidator = createRequiredValidator('Please specify a step between points.')

/**
 * Custom validator for chart series. Since series has complex type
 * we can't validate this with standard validators.
 */
export const insightSeriesValidator: Validator<EditableDataSeries[]> = series => {
    if (!series || series.length === 0) {
        return 'No series defined. You must add at least one series to create a code insight.'
    }

    if (series.some(series => !series.valid)) {
        return 'Some series is invalid. Remove or edit the invalid series.'
    }

    return
}

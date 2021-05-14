import { renderError } from '../../../../../../components/alerts';
import { Validator } from '../../../../../components/form/hooks/useField';
import { AsyncValidator } from '../../../../../components/form/hooks/utils/use-async-validation';
import { createRequiredValidator } from '../../../../../components/form/validators';
import { fetchRepositories } from '../../../../../core/backend/requests/fetch-repositories';
import { DataSeries } from '../../../../../core/backend/types';
import { getSanitizedRepositories } from '../../utils/insight-sanitizer';

export const repositoriesFieldValidator = createRequiredValidator('Repositories is a required field.')
export const requiredStepValueField = createRequiredValidator('Please specify a step between points.')
/**
 * Custom validator for chart series. Since series has complex type
 * we can't validate this with standard validators.
 * */
export const seriesRequired: Validator<DataSeries[]> = series =>
    series && series.length > 0 ? undefined : 'Series is empty. You must have at least one series for code insight.'

export const repositoriesExistValidator: AsyncValidator<string> = async value => {
    if (!value) {
        return;
    }

    try {
        const repositoryNames = getSanitizedRepositories(value);
        const repositories = await fetchRepositories(repositoryNames).toPromise()

        const nullRepositoryIndex = repositories.findIndex(repo => !repo)

        if (nullRepositoryIndex !== -1) {
            return `We couldn't find repository with ${repositoryNames[nullRepositoryIndex]}. Please check the repo URL.`
        }

        return

    } catch (error) {
        return renderError(error);
    }
}

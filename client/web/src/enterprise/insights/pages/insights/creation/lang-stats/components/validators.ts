import { useCallback } from 'react'

import { gql, useLazyQuery } from '@apollo/client'

import { renderError, type AsyncValidator, type Validator } from '@sourcegraph/wildcard'

import type {
    CheckRepositoryExistsResult,
    CheckRepositoryExistsVariables,
} from '../../../../../../../graphql-operations'

/**
 * Primarily used in creation and edit insight pages and also on the landing page where
 * we have a creation UI insight sandbox demo widget.
 */
export const repositoryValidator: Validator<string> = value => {
    if (value !== undefined && value.trim() === '') {
        return 'Repositories is a required field.'
    }

    return
}

const CHECK_REPOSITORY = gql`
    query CheckRepositoryExists($name: String) {
        repository(name: $name) {
            name
        }
    }
`

/**
 * Check that repository exists on the backend. If this repository doesn't exist it
 * will return error validation message.
 */
export function useRepositoryExistsValidator(): AsyncValidator<string> {
    const [checkRepositoryExists] = useLazyQuery<CheckRepositoryExistsResult, CheckRepositoryExistsVariables>(
        CHECK_REPOSITORY,
        {
            fetchPolicy: 'network-only',
        }
    )

    return useCallback(
        async repositoryName => {
            if (!repositoryName || repositoryName.trim() === '') {
                return
            }

            try {
                const { data, error } = await checkRepositoryExists({ variables: { name: repositoryName } })

                if (!data || error || !data.repository?.name) {
                    return `We couldn't find the repository ${repositoryName}. Please ensure the repository exists.`
                }
            } catch (error) {
                return renderError(error)
            }

            return
        },
        [checkRepositoryExists]
    )
}

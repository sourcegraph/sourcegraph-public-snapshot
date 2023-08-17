import { type ApolloClient, gql } from '@apollo/client'

import { InputStatus, type useFieldAPI, type ValidationResult } from '@sourcegraph/wildcard'

import type { GetSearchContextByNameResult } from '../../../../../../../../../graphql-operations'

export const REPO_FILTER_VALIDATORS = isValidRegexp

function isValidRegexp(value = ''): ValidationResult {
    if (value.trim() === '') {
        return
    }

    try {
        new RegExp(value)

        return
    } catch {
        return 'Must be a valid regular expression string'
    }
}

const GET_CONTEXT_BY_NAME = gql`
    query GetSearchContextByName($query: String!) {
        searchContexts(query: $query) {
            nodes {
                spec
                query
            }
        }
    }
`

export const createSearchContextValidator =
    (client: ApolloClient<unknown>) =>
    async (value: string | undefined): Promise<string | void> => {
        if (!value) {
            return
        }

        try {
            const sanitizedValue = value.trim()
            const { data, error } = await client.query<GetSearchContextByNameResult>({
                query: GET_CONTEXT_BY_NAME,
                variables: { query: sanitizedValue },
            })

            if (error) {
                return error.message
            }

            const nodes = data.searchContexts.nodes.filter(node => node.query !== '')

            if (!nodes.some(context => context.spec === sanitizedValue)) {
                return `We couldn't find the context ${sanitizedValue}. Please ensure the context exists.`
            }

            return
        } catch (error) {
            return error.message
        }
    }

export function getFilterInputStatus<T>({ meta }: useFieldAPI<T>): InputStatus {
    const isValidated = meta.initialValue || meta.touched

    if (meta.validState === 'CHECKING') {
        return InputStatus.loading
    }

    if (isValidated && meta.validState === 'VALID') {
        return InputStatus.initial
    }

    if (isValidated && meta.error) {
        return InputStatus.error
    }

    return InputStatus.initial
}

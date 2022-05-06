import { ApolloClient, gql } from '@apollo/client'

import { InputStatus } from '@sourcegraph/wildcard'

import { GetSearchContextByNameResult } from '../../../../../../../../../graphql-operations'
import { useFieldAPI } from '../../../../../../form/hooks/useField'

const GET_CONTEXT_BY_NAME = gql`
    query GetSearchContextByName($query: String!) {
        searchContexts(query: $query) {
            nodes {
                spec
            }
        }
    }
`

export const createSearchContextValidator = (client: ApolloClient<unknown>) => async (
    value: string | undefined
): Promise<string | void> => {
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

        const {
            searchContexts: { nodes },
        } = data

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

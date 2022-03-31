import { ApolloError, ApolloQueryResult } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import { EXTERNAL_SERVICES_WITH_COLLABORATORS } from '../components/externalServices/backend'
import {
    Exact,
    ExternalServicesWithCollaboratorsResult,
    ExternalServicesWithCollaboratorsVariables,
    ListExternalServiceFields,
    ListExternalServiceInvitableCollaboratorsFields,
    Maybe,
} from '../graphql-operations'

interface UseExternalServicesWithCollaboratorsResult {
    externalServices: (ListExternalServiceFields & ListExternalServiceInvitableCollaboratorsFields)[] | undefined
    loadingServices: boolean
    errorServices: ApolloError | undefined
    refetchExternalServices: (
        variables?:
            | Partial<
                  Exact<{
                      first: Maybe<number>
                      after: Maybe<string>
                      namespace: Maybe<string>
                  }>
              >
            | undefined
    ) => Promise<ApolloQueryResult<ExternalServicesWithCollaboratorsResult>>
}

export const useExternalServicesWithCollaborators = (
    userId: string | null
): UseExternalServicesWithCollaboratorsResult => {
    const { data, loading, error, refetch } = useQuery<
        ExternalServicesWithCollaboratorsResult,
        ExternalServicesWithCollaboratorsVariables
    >(EXTERNAL_SERVICES_WITH_COLLABORATORS, {
        variables: {
            namespace: userId,
            first: null,
            after: null,
        },
    })

    return {
        externalServices: data?.externalServices.nodes,
        loadingServices: loading,
        errorServices: error,
        refetchExternalServices: refetch,
    }
}

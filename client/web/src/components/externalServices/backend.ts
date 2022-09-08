import { MutationTuple } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql, dataOrThrowErrors, useMutation } from '@sourcegraph/http-client'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { requestGraphQL } from '../../backend/graphql'
import {
    UpdateExternalServiceResult,
    UpdateExternalServiceVariables,
    Scalars,
    AddExternalServiceVariables,
    AddExternalServiceResult,
    ExternalServiceFields,
    DeleteExternalServiceVariables,
    DeleteExternalServiceResult,
    ExternalServicesVariables,
    ExternalServicesResult,
    AffiliatedRepositoriesVariables,
    AffiliatedRepositoriesResult,
    SyncExternalServiceResult,
    SyncExternalServiceVariables,
    ExternalServiceSyncJobsVariables,
    ExternalServiceSyncJobConnectionFields,
    ExternalServiceSyncJobsResult,
} from '../../graphql-operations'

export const externalServiceFragment = gql`
    fragment ExternalServiceFields on ExternalService {
        id
        kind
        displayName
        config
        warning
        lastSyncError
        repoCount
        webhookURL
        lastSyncAt
        nextSyncAt
        updatedAt
        createdAt
        grantedScopes
        namespace {
            id
            namespaceName
            url
        }
    }
`

export async function addExternalService(
    variables: AddExternalServiceVariables,
    eventLogger: TelemetryService
): Promise<AddExternalServiceResult['addExternalService']> {
    return requestGraphQL<AddExternalServiceResult, AddExternalServiceVariables>(
        gql`
            mutation AddExternalService($input: AddExternalServiceInput!) {
                addExternalService(input: $input) {
                    ...ExternalServiceFields
                }
            }

            ${externalServiceFragment}
        `,
        variables
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.addExternalService || (errors && errors.length > 0)) {
                    eventLogger.log('AddExternalServiceFailed')
                    throw createAggregateError(errors)
                }
                eventLogger.log('AddExternalServiceSucceeded')
                return data.addExternalService
            })
        )
        .toPromise()
}

export const UPDATE_EXTERNAL_SERVICE = gql`
    mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
        updateExternalService(input: $input) {
            ...ExternalServiceFields
        }
    }
    ${externalServiceFragment}
`

export const useUpdateExternalService = (
    onCompleted: (result: UpdateExternalServiceResult) => void
): MutationTuple<UpdateExternalServiceResult, UpdateExternalServiceVariables> =>
    useMutation(UPDATE_EXTERNAL_SERVICE, { onCompleted })

export function updateExternalService(
    variables: UpdateExternalServiceVariables
): Promise<UpdateExternalServiceResult['updateExternalService']> {
    return requestGraphQL<UpdateExternalServiceResult, UpdateExternalServiceVariables>(
        UPDATE_EXTERNAL_SERVICE,
        variables
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.updateExternalService)
        )
        .toPromise()
}

export const FETCH_EXTERNAL_SERVICE = gql`
    query ExternalService($id: ID!) {
        node(id: $id) {
            __typename
            ...ExternalServiceFields
        }
    }
    ${externalServiceFragment}
`

export function listAffiliatedRepositories(
    args: AffiliatedRepositoriesVariables
): Observable<NonNullable<AffiliatedRepositoriesResult>> {
    return requestGraphQL<AffiliatedRepositoriesResult, AffiliatedRepositoriesVariables>(
        gql`
            query AffiliatedRepositories($namespace: ID!, $codeHost: ID, $query: String) {
                affiliatedRepositories(namespace: $namespace, codeHost: $codeHost, query: $query) {
                    nodes {
                        name
                        codeHost {
                            kind
                            id
                            displayName
                        }
                        private
                    }
                    codeHostErrors
                }
            }
        `,
        {
            namespace: args.namespace,
            codeHost: args.codeHost ?? null,
            query: args.query ?? null,
        }
    ).pipe(map(dataOrThrowErrors))
}

export async function deleteExternalService(externalService: Scalars['ID']): Promise<void> {
    const result = await requestGraphQL<DeleteExternalServiceResult, DeleteExternalServiceVariables>(
        gql`
            mutation DeleteExternalService($externalService: ID!) {
                deleteExternalService(externalService: $externalService) {
                    alwaysNil
                }
            }
        `,
        { externalService }
    ).toPromise()
    dataOrThrowErrors(result)
}

export const listExternalServiceFragment = gql`
    fragment ListExternalServiceFields on ExternalService {
        id
        kind
        displayName
        config
        warning
        lastSyncError
        repoCount
        lastSyncAt
        nextSyncAt
        updatedAt
        createdAt
        namespace {
            id
            namespaceName
            url
        }
        grantedScopes
    }
`
export const EXTERNAL_SERVICES = gql`
    query ExternalServices($first: Int, $after: String, $namespace: ID) {
        externalServices(first: $first, after: $after, namespace: $namespace) {
            nodes {
                ...ListExternalServiceFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${listExternalServiceFragment}
`

export function queryExternalServices(
    variables: ExternalServicesVariables
): Observable<ExternalServicesResult['externalServices']> {
    return requestGraphQL<ExternalServicesResult, ExternalServicesVariables>(EXTERNAL_SERVICES, variables).pipe(
        map(({ data, errors }) => {
            if (!data || !data.externalServices || errors) {
                throw createAggregateError(errors)
            }
            return data.externalServices
        })
    )
}

interface ExternalServicesScopeVariables {
    namespace: Scalars['ID']
}

interface ExternalServicesScopeResult {
    externalServices: {
        nodes: {
            id: ExternalServiceFields['id']
            kind: ExternalServiceFields['kind']
            grantedScopes: string[]
        }[]
    }
}

export function queryExternalServicesScope(
    variables: ExternalServicesScopeVariables
): Observable<ExternalServicesScopeResult['externalServices']> {
    return requestGraphQL<ExternalServicesScopeResult, ExternalServicesScopeVariables>(
        gql`
            query ExternalServicesScopes($namespace: ID!) {
                externalServices(first: null, after: null, namespace: $namespace) {
                    nodes {
                        id
                        kind
                        grantedScopes
                    }
                }
            }
        `,
        variables
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.externalServices || errors) {
                throw createAggregateError(errors)
            }
            return data.externalServices
        })
    )
}

export const SYNC_EXTERNAL_SERVICE = gql`
    mutation SyncExternalService($id: ID!) {
        syncExternalService(id: $id) {
            alwaysNil
        }
    }
`

export function useSyncExternalService(): MutationTuple<SyncExternalServiceResult, SyncExternalServiceVariables> {
    return useMutation<SyncExternalServiceResult, SyncExternalServiceVariables>(SYNC_EXTERNAL_SERVICE)
}

export const EXTERNAL_SERVICE_SYNC_JOBS = gql`
    query ExternalServiceSyncJobs($first: Int, $externalService: ID!) {
        node(id: $externalService) {
            __typename
            ... on ExternalService {
                syncJobs(first: $first) {
                    ...ExternalServiceSyncJobConnectionFields
                }
            }
        }
    }

    fragment ExternalServiceSyncJobConnectionFields on ExternalServiceSyncJobConnection {
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
        nodes {
            ...ExternalServiceSyncJobListFields
        }
    }

    fragment ExternalServiceSyncJobListFields on ExternalServiceSyncJob {
        __typename
        id
        state
        startedAt
        finishedAt
        failureMessage
    }
`

export function queryExternalServiceSyncJobs(
    variables: ExternalServiceSyncJobsVariables
): Observable<ExternalServiceSyncJobConnectionFields> {
    return requestGraphQL<ExternalServiceSyncJobsResult, ExternalServiceSyncJobsVariables>(
        EXTERNAL_SERVICE_SYNC_JOBS,
        variables
    ).pipe(
        map(({ data, errors }) => {
            if (errors) {
                throw createAggregateError(errors)
            }
            if (!data) {
                throw new Error('No data found')
            }
            if (!data.node) {
                throw new Error('External service not found')
            }
            if (data.node.__typename !== 'ExternalService') {
                throw new Error(`Node is a ${data.node.__typename}, not ExternalService`)
            }
            return data.node.syncJobs
        })
    )
}

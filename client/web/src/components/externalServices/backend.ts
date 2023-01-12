import { QueryTuple, MutationTuple } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql, dataOrThrowErrors, useMutation, useLazyQuery } from '@sourcegraph/http-client'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { requestGraphQL } from '../../backend/graphql'
import {
    UpdateExternalServiceResult,
    UpdateExternalServiceVariables,
    Scalars,
    AddExternalServiceVariables,
    AddExternalServiceResult,
    DeleteExternalServiceVariables,
    DeleteExternalServiceResult,
    ExternalServicesVariables,
    ExternalServicesResult,
    ExternalServiceCheckConnectionByIdVariables,
    ExternalServiceCheckConnectionByIdResult,
    SyncExternalServiceResult,
    SyncExternalServiceVariables,
    ExternalServiceSyncJobsVariables,
    ExternalServiceSyncJobConnectionFields,
    ExternalServiceSyncJobsResult,
    CancelExternalServiceSyncVariables,
    CancelExternalServiceSyncResult,
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
        lastSyncAt
        nextSyncAt
        updatedAt
        createdAt
        webhookURL
        hasConnectionCheck
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

export const EXTERNAL_SERVICE_CHECK_CONNECTION_BY_ID = gql`
    query ExternalServiceCheckConnectionById($id: ID!) {
        node(id: $id) {
            __typename
            ... on ExternalService {
                id
                hasConnectionCheck
                checkConnection {
                    __typename
                    ... on ExternalServiceUnavailable {
                        suspectedReason
                    }
                }
            }
        }
    }
`

export const useExternalServiceCheckConnectionByIdLazyQuery = (
    id: string
): QueryTuple<ExternalServiceCheckConnectionByIdResult, ExternalServiceCheckConnectionByIdVariables> =>
    useLazyQuery<ExternalServiceCheckConnectionByIdResult, ExternalServiceCheckConnectionByIdVariables>(
        EXTERNAL_SERVICE_CHECK_CONNECTION_BY_ID,
        {
            variables: { id },
        }
    )

const EXTERNAL_SERVICE_SYNC_JOB_LIST_FIELDS_FRAGMENT = gql`
    fragment ExternalServiceSyncJobListFields on ExternalServiceSyncJob {
        __typename
        id
        state
        startedAt
        finishedAt
        failureMessage
        reposSynced
        repoSyncErrors
        reposAdded
        reposDeleted
        reposModified
        reposUnmodified
    }
`

const EXTERNAL_SERVICE_SYNC_JOB_CONNECTION_FIELDS_FRAGMENT = gql`
    ${EXTERNAL_SERVICE_SYNC_JOB_LIST_FIELDS_FRAGMENT}
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
`

export const EXTERNAL_SERVICE_SYNC_JOBS = gql`
    ${EXTERNAL_SERVICE_SYNC_JOB_CONNECTION_FIELDS_FRAGMENT}
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
`

const LIST_EXTERNAL_SERVICE_FRAGMENT = gql`
    ${EXTERNAL_SERVICE_SYNC_JOB_CONNECTION_FIELDS_FRAGMENT}
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
        webhookURL
        hasConnectionCheck
        syncJobs(first: 1) {
            ...ExternalServiceSyncJobConnectionFields
        }
    }
`

export const FETCH_EXTERNAL_SERVICE = gql`
    query ExternalService($id: ID!) {
        node(id: $id) {
            __typename
            ...ListExternalServiceFields
        }
    }
    ${LIST_EXTERNAL_SERVICE_FRAGMENT}
`

export const EXTERNAL_SERVICES = gql`
    query ExternalServices($first: Int, $after: String) {
        externalServices(first: $first, after: $after) {
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

    ${LIST_EXTERNAL_SERVICE_FRAGMENT}
`

export const EXTERNAL_SERVICE_IDS_AND_NAMES = gql`
    query ExternalServiceIDsAndNames {
        externalServices {
            nodes {
                id
                displayName
            }
        }
    }
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

export const CANCEL_EXTERNAL_SERVICE_SYNC = gql`
    mutation CancelExternalServiceSync($id: ID!) {
        cancelExternalServiceSync(id: $id) {
            alwaysNil
        }
    }
`

export function useCancelExternalServiceSync(): MutationTuple<
    CancelExternalServiceSyncResult,
    CancelExternalServiceSyncVariables
> {
    return useMutation<CancelExternalServiceSyncResult, CancelExternalServiceSyncVariables>(
        CANCEL_EXTERNAL_SERVICE_SYNC
    )
}

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

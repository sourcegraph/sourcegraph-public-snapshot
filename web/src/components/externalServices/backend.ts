import { TelemetryService } from '../../../../shared/src/telemetry/telemetryService'
import { Observable } from 'rxjs'
import { gql, dataOrThrowErrors, requestGraphQL } from '../../../../shared/src/graphql/graphql'
import { createAggregateError, isErrorLike, ErrorLike } from '../../../../shared/src/util/errors'
import { map } from 'rxjs/operators'
import {
    UpdateExternalServiceResult,
    UpdateExternalServiceVariables,
    Scalars,
    AddExternalServiceVariables,
    AddExternalServiceResult,
    ExternalServiceFields,
    ExternalServiceVariables,
    ExternalServiceResult,
    DeleteExternalServiceVariables,
    DeleteExternalServiceResult,
    ExternalServicesVariables,
    ExternalServicesResult,
} from '../../graphql-operations'

export const externalServiceFragment = gql`
    fragment ExternalServiceFields on ExternalService {
        id
        kind
        displayName
        config
        warning
        webhookURL
    }
`

export function addExternalService(
    variables: AddExternalServiceVariables,
    eventLogger: TelemetryService
): Observable<AddExternalServiceResult['addExternalService']> {
    return requestGraphQL<AddExternalServiceResult, AddExternalServiceVariables>({
        request: gql`
            mutation AddExternalService($input: AddExternalServiceInput!) {
                addExternalService(input: $input) {
                    ...ExternalServiceFields
                }
            }

            ${externalServiceFragment}
        `,
        variables,
    }).pipe(
        map(({ data, errors }) => {
            if (!data || !data.addExternalService || (errors && errors.length > 0)) {
                eventLogger.log('AddExternalServiceFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('AddExternalServiceSucceeded')
            return data.addExternalService
        })
    )
}

export function isExternalService(
    externalServiceOrError?: ExternalServiceFields | ErrorLike
): externalServiceOrError is ExternalServiceFields {
    return !!externalServiceOrError && !isErrorLike(externalServiceOrError)
}

export function updateExternalService(
    variables: UpdateExternalServiceVariables
): Observable<UpdateExternalServiceResult['updateExternalService']> {
    return requestGraphQL<UpdateExternalServiceResult, UpdateExternalServiceVariables>({
        request: gql`
            mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
                updateExternalService(input: $input) {
                    ...ExternalServiceFields
                }
            }
            ${externalServiceFragment}
        `,
        variables,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateExternalService)
    )
}

export function fetchExternalService(id: Scalars['ID']): Observable<ExternalServiceFields> {
    return requestGraphQL<ExternalServiceResult, ExternalServiceVariables>({
        request: gql`
            query ExternalService($id: ID!) {
                node(id: $id) {
                    __typename
                    ...ExternalServiceFields
                }
            }
            ${externalServiceFragment}
        `,
        variables: { id },
    }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('External service not found')
            }
            if (node.__typename !== 'ExternalService') {
                throw new Error(`Node is a ${node.__typename}, not a ExternalService`)
            }
            return node
        })
    )
}

export async function deleteExternalService(externalService: Scalars['ID']): Promise<void> {
    const result = await requestGraphQL<DeleteExternalServiceResult, DeleteExternalServiceVariables>({
        request: gql`
            mutation DeleteExternalService($externalService: ID!) {
                deleteExternalService(externalService: $externalService) {
                    alwaysNil
                }
            }
        `,
        variables: { externalService },
    }).toPromise()
    dataOrThrowErrors(result)
}

export const listExternalServiceFragment = gql`
    fragment ListExternalServiceFields on ExternalService {
        id
        kind
        displayName
        config
    }
`

export function queryExternalServices(
    variables: ExternalServicesVariables
): Observable<ExternalServicesResult['externalServices']> {
    return requestGraphQL<ExternalServicesResult, ExternalServicesVariables>({
        request: gql`
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
        `,
        variables,
    }).pipe(
        map(({ data, errors }) => {
            if (!data || !data.externalServices || errors) {
                throw createAggregateError(errors)
            }
            return data.externalServices
        })
    )
}

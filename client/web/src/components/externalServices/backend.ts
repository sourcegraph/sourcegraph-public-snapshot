import { Observable } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'

import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { createAggregateError, isErrorLike, ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { requestGraphQL } from '../../backend/graphql'
import {
    UpdateExternalServiceResult,
    UpdateExternalServiceVariables,
    Scalars,
    AddExternalServiceVariables,
    AddExternalServiceResult,
    ExternalServiceFields,
    ExternalServicesVariables,
    ExternalServicesResult,
    SetExternalServiceReposVariables,
    AffiliatedRepositoriesVariables,
    AffiliatedRepositoriesResult,
    AddExternalServiceDocument,
    UpdateExternalServiceDocument,
    SetExternalServiceReposDocument,
    ExternalServiceDocument,
    AffiliatedRepositoriesDocument,
    DeleteExternalServiceDocument,
    ExternalServicesDocument,
    ExternalServicesScopesDocument,
    ExternalServicesScopesResult,
    ExternalServicesScopesVariables,
} from '../../graphql-operations'

export async function addExternalService(
    variables: AddExternalServiceVariables,
    eventLogger: TelemetryService
): Promise<AddExternalServiceResult['addExternalService']> {
    return requestGraphQL(AddExternalServiceDocument, variables)
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

export function isExternalService(
    externalServiceOrError?: ExternalServiceFields | ErrorLike
): externalServiceOrError is ExternalServiceFields {
    return !!externalServiceOrError && !isErrorLike(externalServiceOrError)
}

export function updateExternalService(
    variables: UpdateExternalServiceVariables
): Promise<UpdateExternalServiceResult['updateExternalService']> {
    return requestGraphQL(UpdateExternalServiceDocument, variables)
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.updateExternalService)
        )
        .toPromise()
}

export function setExternalServiceRepos(variables: SetExternalServiceReposVariables): Promise<void> {
    return requestGraphQL(SetExternalServiceReposDocument, variables)
        .pipe(map(dataOrThrowErrors), mapTo(undefined))
        .toPromise()
}

export function fetchExternalService(id: Scalars['ID']): Observable<ExternalServiceFields> {
    return requestGraphQL(ExternalServiceDocument, { id }).pipe(
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

export function listAffiliatedRepositories(
    args: AffiliatedRepositoriesVariables
): Observable<NonNullable<AffiliatedRepositoriesResult>> {
    return requestGraphQL(AffiliatedRepositoriesDocument, {
        user: args.user,
        codeHost: args.codeHost ?? null,
        query: args.query ?? null,
    }).pipe(map(dataOrThrowErrors))
}

export async function deleteExternalService(externalService: Scalars['ID']): Promise<void> {
    const result = await requestGraphQL(DeleteExternalServiceDocument, { externalService }).toPromise()
    dataOrThrowErrors(result)
}

export function queryExternalServices(
    variables: ExternalServicesVariables
): Observable<ExternalServicesResult['externalServices']> {
    return requestGraphQL(ExternalServicesDocument, variables).pipe(
        map(({ data, errors }) => {
            if (!data || !data.externalServices || errors) {
                throw createAggregateError(errors)
            }
            return data.externalServices
        })
    )
}

export function queryExternalServicesScope(
    variables: ExternalServicesScopesVariables
): Observable<ExternalServicesScopesResult['externalServices']> {
    return requestGraphQL(ExternalServicesScopesDocument, variables).pipe(
        map(({ data, errors }) => {
            if (!data || !data.externalServices || errors) {
                throw createAggregateError(errors)
            }
            return data.externalServices
        })
    )
}

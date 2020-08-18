import { parse as parseJSONC } from '@sqs/jsonc-parser'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useEffect, useState, useCallback } from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, Observable } from 'rxjs'
import { catchError, startWith, switchMap, map } from 'rxjs/operators'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { PageTitle } from '../../../components/PageTitle'
import { ExternalServiceCard } from '../../../components/ExternalServiceCard'
import { ExternalServiceForm } from './ExternalServiceForm'
import { ErrorAlert } from '../../../components/alerts'
import { defaultExternalServices, codeHostExternalServices } from '../../../site-admin/externalServices'
import { hasProperty } from '../../../../../shared/src/util/types'
import * as H from 'history'
import { useEventObservable } from '../../../../../shared/src/util/useObservable'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { mutateGraphQL, queryGraphQL } from '../../../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'

type ExternalService = Pick<GQL.IExternalService, 'id' | 'kind' | 'displayName' | 'config' | 'warning' | 'webhookURL'>

interface Props extends RouteComponentProps<{ id: GQL.ID }>, TelemetryProps {
    isLightTheme: boolean
    history: H.History
}

const LOADING = 'loading' as const

export const ExternalServicePage: React.FunctionComponent<Props> = ({
    match,
    history,
    isLightTheme,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('ExternalService')
    }, [telemetryService])

    const [externalServiceOrError, setExternalServiceOrError] = useState<typeof LOADING | ExternalService | ErrorLike>(
        LOADING
    )

    useEffect(() => {
        const subscription = fetchExternalService(match.params.id)
            .pipe(
                startWith(LOADING),
                catchError(error => [asError(error)])
            )
            .subscribe(result => {
                setExternalServiceOrError(result)
            })
        return () => subscription.unsubscribe()
    }, [match.params.id])

    const onChange = useCallback(
        (input: GQL.IAddExternalServiceInput) => {
            if (isExternalService(externalServiceOrError)) {
                setExternalServiceOrError({ ...externalServiceOrError, ...input })
            }
        },
        [externalServiceOrError, setExternalServiceOrError]
    )

    const [nextSubmit, updatedServiceOrError] = useEventObservable(
        useCallback(
            (submits: Observable<GQL.IExternalService>): Observable<typeof LOADING | ErrorLike | ExternalService> =>
                submits.pipe(
                    switchMap(input =>
                        concat(
                            [LOADING],
                            updateExternalService(input).pipe(catchError((error: Error) => [asError(error)]))
                        )
                    )
                ),
            []
        )
    )

    // If the update was successful, and did not surface a warning, redirect to the
    // repositories page, adding `?repositoriesUpdated` to the query string so that we display
    // a banner at the top of the page.
    useEffect(() => {
        if (updatedServiceOrError && updatedServiceOrError !== LOADING && !isErrorLike(updatedServiceOrError)) {
            if (updatedServiceOrError.warning) {
                setExternalServiceOrError(updatedServiceOrError)
            } else {
                history.push(`./${updatedServiceOrError.id}`)
            }
        }
    }, [updatedServiceOrError, history])

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            if (event) {
                event.preventDefault()
            }
            if (isExternalService(externalServiceOrError)) {
                nextSubmit(externalServiceOrError)
            }
        },
        [externalServiceOrError, nextSubmit]
    )
    let error: ErrorLike | undefined
    if (isErrorLike(updatedServiceOrError)) {
        error = updatedServiceOrError
    }

    const externalService =
        (!isErrorLike(externalServiceOrError) && externalServiceOrError !== LOADING && externalServiceOrError) ||
        undefined

    let externalServiceCategory = externalService && defaultExternalServices[externalService.kind]
    if (
        externalService &&
        [GQL.ExternalServiceKind.GITHUB, GQL.ExternalServiceKind.GITLAB].includes(externalService.kind)
    ) {
        const parsedConfig: unknown = parseJSONC(externalService.config)
        const url =
            typeof parsedConfig === 'object' &&
            parsedConfig !== null &&
            hasProperty('url')(parsedConfig) &&
            typeof parsedConfig.url === 'string'
                ? new URL(parsedConfig.url)
                : undefined
        // We have no way of finding out whether a externalservice of kind GITHUB is GitHub.com or GitHub enterprise, so we need to guess based on the URL.
        if (externalService.kind === GQL.ExternalServiceKind.GITHUB && url?.hostname !== 'github.com') {
            externalServiceCategory = codeHostExternalServices.ghe
        }
        // We have no way of finding out whether a externalservice of kind GITLAB is Gitlab.com or Gitlab self-hosted, so we need to guess based on the URL.
        if (externalService.kind === GQL.ExternalServiceKind.GITLAB && url?.hostname !== 'gitlab.com') {
            externalServiceCategory = codeHostExternalServices.gitlab
        }
    }

    return (
        <div className="site-admin-configuration-page">
            {externalService ? (
                <PageTitle title={`External service - ${externalService.displayName}`} />
            ) : (
                <PageTitle title="External service" />
            )}
            <h2>Update synced repositories</h2>
            {externalServiceOrError === LOADING && <LoadingSpinner className="icon-inline" />}
            {isErrorLike(externalServiceOrError) && (
                <ErrorAlert className="mb-3" error={externalServiceOrError} history={history} />
            )}
            {externalServiceCategory && (
                <div className="mb-3">
                    <ExternalServiceCard {...externalServiceCategory} />
                </div>
            )}
            {externalService && externalServiceCategory && (
                <ExternalServiceForm
                    input={externalService}
                    editorActions={externalServiceCategory.editorActions}
                    jsonSchema={externalServiceCategory.jsonSchema}
                    error={error}
                    warning={externalService.warning}
                    mode="edit"
                    loading={updatedServiceOrError === LOADING}
                    onSubmit={onSubmit}
                    onChange={onChange}
                    history={history}
                    isLightTheme={isLightTheme}
                    telemetryService={telemetryService}
                />
            )}
            {externalService && <ExternalServiceWebhook externalService={externalService} />}
        </div>
    )
}

function isExternalService(
    externalServiceOrError: typeof LOADING | ExternalService | ErrorLike
): externalServiceOrError is GQL.IExternalService {
    return externalServiceOrError !== LOADING && !isErrorLike(externalServiceOrError)
}

const externalServiceFragment = gql`
    fragment externalServiceFields on ExternalService {
        id
        kind
        displayName
        config
        warning
        webhookURL
    }
`

function updateExternalService(input: GQL.IUpdateExternalServiceInput): Observable<ExternalService> {
    return mutateGraphQL(
        gql`
            mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
                updateExternalService(input: $input) {
                    ...externalServiceFields
                }
            }
            ${externalServiceFragment}
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateExternalService)
    )
}

function fetchExternalService(id: GQL.ID): Observable<ExternalService> {
    return queryGraphQL(
        gql`
            query ExternalService($id: ID!) {
                node(id: $id) {
                    __typename
                    ...externalServiceFields
                }
            }
            ${externalServiceFragment}
        `,
        { id }
    ).pipe(
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

import * as H from 'history'
import React, { useEffect, useCallback, useState } from 'react'
import { Observable, concat } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'
import { Markdown } from '../../../../shared/src/components/Markdown'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike, asError, isErrorLike } from '../../../../shared/src/util/errors'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { PageTitle } from '../PageTitle'
import { refreshSiteFlags } from '../../site/backend'
import { ThemeProps } from '../../../../shared/src/theme'
import { ExternalServiceCard } from './ExternalServiceCard'
import { AddExternalServiceOptions } from './externalServices'
import { useEventObservable } from '../../../../shared/src/util/useObservable'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { ExternalServiceForm } from './ExternalServiceForm'
import { addExternalService } from './backend'
import { ExternalServiceFields, Scalars } from '../../graphql-operations'

interface Props extends ThemeProps, TelemetryProps {
    history: H.History
    externalService: AddExternalServiceOptions
    routingPrefix: string
    afterCreateRoute: string
    userID?: Scalars['ID']
}

/**
 * Page for adding a single external service.
 */
export const AddExternalServicePage: React.FunctionComponent<Props> = ({
    afterCreateRoute,
    externalService,
    history,
    isLightTheme,
    routingPrefix,
    telemetryService,
    userID,
}) => {
    const [config, setConfig] = useState(externalService.defaultConfig)
    const [displayName, setDisplayName] = useState(externalService.defaultDisplayName)

    useEffect(() => {
        telemetryService.logViewEvent('AddExternalService')
    }, [telemetryService])

    const [nextSubmit, createdServiceOrError] = useEventObservable(
        useCallback(
            (submits: Observable<GQL.IAddExternalServiceInput>): Observable<ErrorLike | ExternalServiceFields> =>
                submits.pipe(
                    switchMap(input =>
                        concat(
                            addExternalService(
                                { input: { ...input, namespace: userID ?? null } },
                                telemetryService
                            ).pipe(catchError(error => [asError(error)]))
                        )
                    )
                ),
            [telemetryService, userID]
        )
    )

    useEffect(() => {
        if (createdServiceOrError && !isErrorLike(createdServiceOrError)) {
            // Refresh site flags so that global site alerts
            // reflect the latest configuration.
            // eslint-disable-next-line rxjs/no-ignored-subscription
            refreshSiteFlags().subscribe({ error: error => console.error(error) })
            history.push(afterCreateRoute)
        }
    }, [createdServiceOrError, history, afterCreateRoute])

    const getExternalServiceInput = useCallback(
        (): GQL.IAddExternalServiceInput => ({
            displayName,
            config,
            kind: externalService.kind,
        }),
        [displayName, config, externalService.kind]
    )

    const onChange = useCallback(
        (input: GQL.IAddExternalServiceInput): void => {
            setDisplayName(input.displayName)
            setConfig(input.config)
        },
        [setDisplayName, setConfig]
    )

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            if (event) {
                event.preventDefault()
            }
            nextSubmit(getExternalServiceInput())
        },
        [nextSubmit, getExternalServiceInput]
    )

    return (
        <div className="add-external-service-page mt-3">
            <PageTitle title="Add repositories" />
            <h2>Add repositories</h2>
            {createdServiceOrError && !isErrorLike(createdServiceOrError) && createdServiceOrError.warning ? (
                <div>
                    <div className="mb-3">
                        <ExternalServiceCard
                            {...externalService}
                            title={createdServiceOrError.displayName}
                            shortDescription="Update this external service configuration to manage repository mirroring."
                            to={`${routingPrefix}/external-services/${createdServiceOrError.id}`}
                        />
                    </div>
                    <div className="alert alert-warning">
                        <h4>Warning</h4>
                        <Markdown
                            dangerousInnerHTML={renderMarkdown(createdServiceOrError.warning)}
                            history={history}
                        />
                    </div>
                </div>
            ) : (
                <div>
                    <div className="mb-3">
                        <ExternalServiceCard {...externalService} />
                    </div>
                    <h3>Instructions:</h3>
                    <div className="mb-4">{externalService.instructions}</div>
                    <ExternalServiceForm
                        history={history}
                        isLightTheme={isLightTheme}
                        telemetryService={telemetryService}
                        error={isErrorLike(createdServiceOrError) ? createdServiceOrError : undefined}
                        input={getExternalServiceInput()}
                        editorActions={externalService.editorActions}
                        jsonSchema={externalService.jsonSchema}
                        mode="create"
                        onSubmit={onSubmit}
                        onChange={onChange}
                        loading={createdServiceOrError === undefined}
                    />
                </div>
            )}
        </div>
    )
}

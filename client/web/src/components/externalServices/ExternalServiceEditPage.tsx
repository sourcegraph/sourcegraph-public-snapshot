import React, { useEffect, useState, useCallback, useMemo } from 'react'

import * as H from 'history'
import { parse as parseJSONC } from 'jsonc-parser'
import { Redirect } from 'react-router'
import { Subject } from 'rxjs'

import { hasProperty } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, H2, Container, ErrorAlert } from '@sourcegraph/wildcard'

import {
    ExternalServiceFields,
    Scalars,
    AddExternalServiceInput,
    ExternalServiceResult,
    ExternalServiceVariables,
    ExternalServiceKind,
} from '../../graphql-operations'
import { LoaderButton } from '../LoaderButton'
import { PageTitle } from '../PageTitle'

import {
    useSyncExternalService,
    queryExternalServiceSyncJobs as _queryExternalServiceSyncJobs,
    useUpdateExternalService,
    FETCH_EXTERNAL_SERVICE,
} from './backend'
import { ExternalServiceCard } from './ExternalServiceCard'
import { ExternalServiceForm } from './ExternalServiceForm'
import { defaultExternalServices, codeHostExternalServices } from './externalServices'
import { ExternalServiceSyncJobsList } from './ExternalServiceSyncJobsList'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'

interface Props extends TelemetryProps {
    externalServiceID: Scalars['ID']
    isLightTheme: boolean
    history: H.History
    afterUpdateRoute: string

    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean

    /** For testing only. */
    queryExternalServiceSyncJobs?: typeof _queryExternalServiceSyncJobs
    /** For testing only. */
    autoFocusForm?: boolean
}

function isValidURL(url: string): boolean {
    try {
        new URL(url)
        return true
    } catch {
        return false
    }
}

const getExternalService = (queryResult?: ExternalServiceResult): ExternalServiceFields | null =>
    queryResult?.node?.__typename === 'ExternalService' ? queryResult.node : null

export const ExternalServiceEditPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalServiceID,
    history,
    isLightTheme,
    telemetryService,
    afterUpdateRoute,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
    queryExternalServiceSyncJobs = _queryExternalServiceSyncJobs,
    autoFocusForm,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalService')
    }, [telemetryService])

    const [externalService, setExternalService] = useState<ExternalServiceFields>()

    const { error: fetchError, loading: fetchLoading } = useQuery<ExternalServiceResult, ExternalServiceVariables>(
        FETCH_EXTERNAL_SERVICE,
        {
            variables: { id: externalServiceID },
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
            onCompleted: result => {
                const data = getExternalService(result)
                if (data) {
                    setExternalService(data)
                }
            },
        }
    )

    const [syncExternalService, { error: syncExternalServiceError, loading: syncExternalServiceLoading }] =
        useSyncExternalService()

    const [updated, setUpdated] = useState(false)
    const [updateExternalService, { error: updateExternalServiceError, loading: updateExternalServiceLoading }] =
        useUpdateExternalService(result => {
            setExternalService(result.updateExternalService)
            setUpdated(true)
        })

    const onSubmit = useCallback(
        async (event?: React.FormEvent<HTMLFormElement>) => {
            event?.preventDefault()

            if (externalService !== undefined) {
                await updateExternalService({
                    variables: {
                        input: {
                            id: externalService.id,
                            displayName: externalService.displayName,
                            config: externalService.config,
                        },
                    },
                })
            }
        },
        [externalService, updateExternalService]
    )

    const onChange = useCallback(
        (input: AddExternalServiceInput) => {
            if (externalService !== undefined) {
                setExternalService({
                    ...externalService,
                    ...input,
                })
            }
        },
        [externalService, setExternalService]
    )

    const syncJobUpdates = useMemo(() => new Subject<void>(), [])
    const triggerSync = useCallback(
        () =>
            externalService &&
            syncExternalService({ variables: { id: externalService.id } }).then(() => {
                syncJobUpdates.next()
            }),
        [externalService, syncExternalService, syncJobUpdates]
    )

    let externalServiceCategory = externalService && defaultExternalServices[externalService.kind]
    if (externalService && [ExternalServiceKind.GITHUB, ExternalServiceKind.GITLAB].includes(externalService.kind)) {
        const parsedConfig: unknown = parseJSONC(externalService.config)
        const url =
            typeof parsedConfig === 'object' &&
            parsedConfig !== null &&
            hasProperty('url')(parsedConfig) &&
            typeof parsedConfig.url === 'string' &&
            isValidURL(parsedConfig.url)
                ? new URL(parsedConfig.url)
                : undefined
        // We have no way of finding out whether an external service is GITHUB or GitHub.com or GitHub enterprise, so we need to guess based on the URL.
        if (externalService.kind === ExternalServiceKind.GITHUB && url?.hostname !== 'github.com') {
            externalServiceCategory = codeHostExternalServices.ghe
        }
        // We have no way of finding out whether an external service is GITLAB or Gitlab.com or Gitlab self-hosted, so we need to guess based on the URL.
        if (externalService.kind === ExternalServiceKind.GITLAB && url?.hostname !== 'gitlab.com') {
            externalServiceCategory = codeHostExternalServices.gitlab
        }
    }

    const combinedError = fetchError || updateExternalServiceError
    const combinedLoading = fetchLoading || updateExternalServiceLoading

    if (updated && !combinedLoading && externalService?.warning === null) {
        return <Redirect to={afterUpdateRoute} />
    }

    return (
        <div>
            {externalService ? (
                <PageTitle title={`Code host - ${externalService.displayName}`} />
            ) : (
                <PageTitle title="Code host" />
            )}
            <H2>Update code host connection {combinedLoading && <LoadingSpinner inline={true} />}</H2>
            {combinedError !== undefined && !combinedLoading && <ErrorAlert className="mb-3" error={combinedError} />}

            {externalService && (
                <Container className="mb-3">
                    {externalServiceCategory && (
                        <div className="mb-3">
                            <ExternalServiceCard {...externalServiceCategory} />
                        </div>
                    )}
                    {externalServiceCategory && (
                        <ExternalServiceForm
                            input={{ ...externalService }}
                            editorActions={externalServiceCategory.editorActions}
                            jsonSchema={externalServiceCategory.jsonSchema}
                            error={updateExternalServiceError}
                            warning={externalService.warning}
                            mode="edit"
                            loading={combinedLoading}
                            onSubmit={onSubmit}
                            onChange={onChange}
                            history={history}
                            isLightTheme={isLightTheme}
                            telemetryService={telemetryService}
                            autoFocus={autoFocusForm}
                            externalServicesFromFile={externalServicesFromFile}
                            allowEditExternalServicesWithFile={allowEditExternalServicesWithFile}
                        />
                    )}
                    <LoaderButton
                        label="Trigger manual sync"
                        alwaysShowLabel={true}
                        variant="secondary"
                        onClick={triggerSync}
                        loading={syncExternalServiceLoading}
                        disabled={syncExternalServiceLoading}
                    />
                    {syncExternalServiceError && <ErrorAlert error={syncExternalServiceError} />}
                    <ExternalServiceWebhook externalService={externalService} className="mt-3" />
                    <ExternalServiceSyncJobsList
                        queryExternalServiceSyncJobs={queryExternalServiceSyncJobs}
                        externalServiceID={externalService.id}
                        updates={syncJobUpdates}
                    />
                </Container>
            )}
        </div>
    )
}

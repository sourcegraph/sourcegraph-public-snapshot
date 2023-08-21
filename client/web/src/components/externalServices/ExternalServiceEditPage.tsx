import React, { type FC, useEffect, useState, useCallback, useMemo } from 'react'

import { useNavigate, useParams } from 'react-router-dom'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, PageHeader, ButtonLink } from '@sourcegraph/wildcard'

import type { AddExternalServiceInput } from '../../graphql-operations'
import { CreatedByAndUpdatedByInfoByline } from '../Byline/CreatedByAndUpdatedByInfoByline'
import { useFetchGithubAppForES } from '../gitHubApps/backend'
import { PageTitle } from '../PageTitle'

import {
    useUpdateExternalService,
    type ExternalServiceFieldsWithConfig,
    useFetchExternalService,
    getExternalService,
} from './backend'
import { getBreadCrumbs } from './breadCrumbs'
import { ExternalServiceForm } from './ExternalServiceForm'
import { resolveExternalServiceCategory } from './externalServices'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'

interface Props extends TelemetryProps {
    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean

    /** For testing only. */
    autoFocusForm?: boolean
}

export const ExternalServiceEditPage: FC<Props> = ({
    telemetryService,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
    autoFocusForm,
}) => {
    const { externalServiceID } = useParams()
    const navigate = useNavigate()

    useEffect(() => {
        telemetryService.logViewEvent('SiteAdminExternalService')
    }, [telemetryService])

    const [externalService, setExternalService] = useState<ExternalServiceFieldsWithConfig>()

    const { error: fetchError, loading: fetchLoading } = useFetchExternalService(externalServiceID!, setExternalService)

    const [updated, setUpdated] = useState(false)
    const [updateExternalService, { error: updateExternalServiceError, loading: updateExternalServiceLoading }] =
        useUpdateExternalService(result => {
            setExternalService(getExternalService(result.updateExternalService))
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
                setExternalService(
                    getExternalService({
                        ...externalService,
                        ...input,
                    })
                )
            }
        },
        [externalService, setExternalService]
    )

    const path = useMemo(() => getBreadCrumbs(externalService, true), [externalService])

    const combinedError = fetchError || updateExternalServiceError
    const combinedLoading = fetchLoading || updateExternalServiceLoading

    if (updated && !combinedLoading && externalService?.warning === null) {
        navigate(`/site-admin/external-services/${encodeURIComponent(externalService.id)}`, { replace: true })
    }

    const ExternalServiceContainer: FC<{ externalService: ExternalServiceFieldsWithConfig }> = ({
        externalService,
    }) => {
        const {
            error: fetchGHAppError,
            loading: fetchGHAppLoading,
            data: ghAppData,
        } = useFetchGithubAppForES(externalService)
        const ghApp = useMemo(() => ghAppData?.gitHubAppByAppID, [ghAppData])
        const externalServiceCategory = useMemo(
            () => resolveExternalServiceCategory(externalService, ghApp),
            [externalService, ghApp]
        )
        return (
            <Container className="mb-3">
                <PageHeader
                    path={path}
                    byline={
                        <CreatedByAndUpdatedByInfoByline
                            createdAt={externalService.createdAt}
                            updatedAt={externalService.updatedAt}
                            noAuthor={true}
                        />
                    }
                    className="mb-3"
                    headingElement="h2"
                    actions={
                        <ButtonLink
                            to={`/site-admin/external-services/${encodeURIComponent(externalService.id)}`}
                            variant="secondary"
                        >
                            Cancel
                        </ButtonLink>
                    }
                />
                {fetchGHAppError !== undefined && !fetchGHAppLoading && (
                    <ErrorAlert className="mb-3" error={fetchGHAppError} />
                )}
                {externalServiceCategory && (
                    <ExternalServiceForm
                        input={{ ...externalService }}
                        externalServiceID={externalServiceID}
                        editorActions={externalServiceCategory.editorActions}
                        jsonSchema={externalServiceCategory.jsonSchema}
                        error={updateExternalServiceError}
                        warning={externalService.warning}
                        mode="edit"
                        loading={combinedLoading}
                        onSubmit={onSubmit}
                        onChange={onChange}
                        telemetryService={telemetryService}
                        autoFocus={autoFocusForm}
                        externalServicesFromFile={externalServicesFromFile}
                        allowEditExternalServicesWithFile={allowEditExternalServicesWithFile}
                        additionalFormComponent={externalServiceCategory.additionalFormComponent}
                    />
                )}
                <ExternalServiceWebhook externalService={externalService} className="mt-3" />
            </Container>
        )
    }

    return (
        <div>
            {externalService ? (
                <PageTitle title={`Code host - ${externalService.displayName}`} />
            ) : (
                <PageTitle title="Code host" />
            )}
            {combinedError !== undefined && !combinedLoading && <ErrorAlert className="mb-3" error={combinedError} />}
            {externalService && <ExternalServiceContainer externalService={externalService} />}
        </div>
    )
}

import React, { FC, useEffect, useState, useCallback, useMemo } from 'react'

import { Link, Navigate, useNavigate, useParams } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, PageHeader, Button, Label } from '@sourcegraph/wildcard'

import { AddExternalServiceInput } from '../../graphql-operations'
import { CreatedByAndUpdatedByInfoByline } from '../Byline/CreatedByAndUpdatedByInfoByline'
import { useFetchGithubAppForES } from '../gitHubApps/backend'
import { PageTitle } from '../PageTitle'

import {
    useUpdateExternalService,
    ExternalServiceFieldsWithConfig,
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
    const {
        error: fetchGHAppError,
        loading: fetchGHAppLoading,
        data: ghAppData,
    } = useFetchGithubAppForES(externalService)
    const ghApp = useMemo(() => ghAppData?.gitHubAppByAppID, [ghAppData])

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

    const externalServiceCategory = resolveExternalServiceCategory(externalService)

    const combinedError = fetchError || updateExternalServiceError || fetchGHAppError
    const combinedLoading = fetchLoading || updateExternalServiceLoading || fetchGHAppLoading

    if (updated && !combinedLoading && externalService?.warning === null) {
        return <Navigate to={`/site-admin/external-services/${externalService.id}`} replace={true} />
    }

    return (
        <div>
            {externalService ? (
                <PageTitle title={`Code host - ${externalService.displayName}`} />
            ) : (
                <PageTitle title="Code host" />
            )}
            {combinedError !== undefined && !combinedLoading && <ErrorAlert className="mb-3" error={combinedError} />}
            {externalService && (
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
                            <Button onClick={() => navigate(-1)} variant="secondary">
                                Cancel
                            </Button>
                        }
                    />
                    {ghApp && (
                        <Label>
                            GitHub App:
                            <Link className="ml-3" to={`/site-admin/github-apps/${ghApp.id}`}>
                                {ghApp.name}
                            </Link>
                        </Label>
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
                        />
                    )}
                    <ExternalServiceWebhook externalService={externalService} className="mt-3" />
                </Container>
            )}
        </div>
    )
}

import React, { useEffect, useState, useCallback } from 'react'

import { mdiCog } from '@mdi/js'
import * as H from 'history'
import { Redirect } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, PageHeader, Icon, ButtonLink } from '@sourcegraph/wildcard'

import {
    ExternalServiceFields,
    Scalars,
    AddExternalServiceInput,
    ExternalServiceResult,
    ExternalServiceVariables,
} from '../../graphql-operations'
import { CreatedByAndUpdatedByInfoByline } from '../Byline/CreatedByAndUpdatedByInfoByline'
import { PageTitle } from '../PageTitle'

import { useUpdateExternalService, FETCH_EXTERNAL_SERVICE } from './backend'
import { ExternalServiceForm } from './ExternalServiceForm'
import { resolveExternalServiceCategory } from './externalServices'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'

interface Props extends TelemetryProps {
    externalServiceID: Scalars['ID']
    isLightTheme: boolean
    history: H.History
    routingPrefix: string

    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean

    /** For testing only. */
    autoFocusForm?: boolean
}

const getExternalService = (queryResult?: ExternalServiceResult): ExternalServiceFields | null =>
    queryResult?.node?.__typename === 'ExternalService' ? queryResult.node : null

export const ExternalServiceEditPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    externalServiceID,
    history,
    routingPrefix,
    isLightTheme,
    telemetryService,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
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

    const externalServiceCategory = resolveExternalServiceCategory(externalService)

    const combinedError = fetchError || updateExternalServiceError
    const combinedLoading = fetchLoading || updateExternalServiceLoading

    if (updated && !combinedLoading && externalService?.warning === null) {
        return <Redirect to={`${routingPrefix}/external-services/${externalService.id}`} />
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
                        path={[
                            { icon: mdiCog },
                            { to: '/site-admin/external-services', text: 'Code hosts' },
                            {
                                text: (
                                    <>
                                        {externalService.displayName}
                                        {externalServiceCategory && (
                                            <Icon
                                                inline={true}
                                                as={externalServiceCategory.icon}
                                                aria-label="Code host logo"
                                                className="ml-2"
                                            />
                                        )}
                                    </>
                                ),
                            },
                        ]}
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
                            <ButtonLink to={`/site-admin/external-services/${externalServiceID}`} variant="secondary">
                                Cancel
                            </ButtonLink>
                        }
                    />
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
                            history={history}
                            isLightTheme={isLightTheme}
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

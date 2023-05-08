import { FC, useEffect, useCallback, useState } from 'react'

import { FetchResult } from '@apollo/client'
import { useNavigate } from 'react-router-dom'

import { logger, renderMarkdown } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Container, H2, H3, H4, Markdown } from '@sourcegraph/wildcard'

import { AddExternalServiceInput, AddExternalServiceResult } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'
import { PageTitle } from '../PageTitle'

import { useAddExternalService } from './backend'
import { ExternalServiceCard } from './ExternalServiceCard'
import { ExternalServiceForm } from './ExternalServiceForm'
import { AddExternalServiceOptions } from './externalServices'

interface Props extends TelemetryProps {
    externalService: AddExternalServiceOptions
    externalServicesFromFile: boolean
    allowEditExternalServicesWithFile: boolean

    /** For testing only. */
    autoFocusForm?: boolean
}

/**
 * Page for adding a single external service.
 */
export const AddExternalServicePage: FC<Props> = ({
    externalService,
    telemetryService,
    autoFocusForm,
    externalServicesFromFile,
    allowEditExternalServicesWithFile,
}) => {
    const [config, setConfig] = useState(externalService.defaultConfig)
    const [displayName, setDisplayName] = useState(externalService.defaultDisplayName)
    const navigate = useNavigate()
    const { Instructions } = externalService

    useEffect(() => {
        telemetryService.logPageView('AddExternalService')
    }, [telemetryService])

    const getExternalServiceInput = useCallback(
        (): AddExternalServiceInput => ({
            displayName,
            config,
            kind: externalService.kind,
        }),
        [displayName, config, externalService.kind]
    )

    const onChange = useCallback(
        (input: AddExternalServiceInput): void => {
            setDisplayName(input.displayName)
            setConfig(input.config)
        },
        [setDisplayName, setConfig]
    )

    const [addExternalService, { data: addExternalServiceResult, loading: isCreating, error, client }] =
        useAddExternalService()

    const onSubmit = useCallback(
        async (event?: React.FormEvent<HTMLFormElement>): Promise<FetchResult<AddExternalServiceResult>> => {
            if (event) {
                event.preventDefault()
            }
            return addExternalService({
                variables: {
                    input: { ...getExternalServiceInput() },
                },
                onCompleted: data => {
                    telemetryService.log('AddExternalServiceSucceeded')
                    refreshSiteFlags(client).catch((error: Error) => logger.error(error))
                    navigate(`/site-admin/external-services/${data.addExternalService.id}`)
                },
                onError: () => {
                    telemetryService.log('AddExternalServiceFailed')
                },
            })
        },
        [addExternalService, telemetryService, getExternalServiceInput, client, navigate]
    )
    const createdExternalService = addExternalServiceResult?.addExternalService

    return (
        <>
            <PageTitle title="Add code host connection" />
            <H2>Add code host connection</H2>
            <Container>
                {createdExternalService?.warning ? (
                    <div>
                        <div className="mb-3">
                            <ExternalServiceCard
                                {...externalService}
                                title={createdExternalService.displayName}
                                shortDescription="Update this external service configuration to manage repository mirroring."
                                to={`/site-admin/external-services/${createdExternalService.id}/edit`}
                            />
                        </div>
                        <Alert variant="warning">
                            <H4>Warning</H4>
                            <Markdown dangerousInnerHTML={renderMarkdown(createdExternalService.warning)} />
                        </Alert>
                    </div>
                ) : (
                    <>
                        <div className="mb-3">
                            <ExternalServiceCard {...externalService} />
                        </div>
                        {Instructions && (
                            <>
                                <H3>Instructions:</H3>
                                <div className="mb-4">
                                    <Instructions />
                                </div>
                            </>
                        )}
                        <ExternalServiceForm
                            telemetryService={telemetryService}
                            error={error}
                            input={getExternalServiceInput()}
                            editorActions={externalService.editorActions}
                            jsonSchema={externalService.jsonSchema}
                            mode="create"
                            onSubmit={onSubmit}
                            onChange={onChange}
                            loading={isCreating === true}
                            autoFocus={autoFocusForm}
                            externalServicesFromFile={externalServicesFromFile}
                            allowEditExternalServicesWithFile={allowEditExternalServicesWithFile}
                        />
                    </>
                )}
            </Container>
        </>
    )
}

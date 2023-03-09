import React, { useCallback } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, ErrorAlert, screenReaderAnnounce, Text } from '@sourcegraph/wildcard'

import { InferAutoIndexJobsForRepoResult, InferAutoIndexJobsForRepoVariables } from '../../../../graphql-operations'

import { InferenceForm } from './inference-form/InferenceForm'
import { SchemaCompatibleInferenceFormData } from './inference-form/types'
import { INFER_JOBS_SCRIPT } from '../backend'
import { useUpdateConfigurationForRepository } from '../hooks/useUpdateConfigurationForRepository'
import { useInferredConfig } from '../hooks/useInferredConfig'
import { useRepositoryConfig } from '../hooks/useRepositoryConfig'

interface ConfigurationFormProps extends TelemetryProps {
    repoId: string
    authenticatedUser: AuthenticatedUser | null
}

export const ConfigurationForm: React.FunctionComponent<ConfigurationFormProps> = ({ repoId, authenticatedUser }) => {
    const { updateConfigForRepository, isUpdating, updatingError } = useUpdateConfigurationForRepository()
    const { inferredConfiguration, loadingInferred, inferredError } = useInferredConfig(repoId)
    const { configuration, loadingRepository, repositoryError } = useRepositoryConfig(repoId)

    // Use the available configuration if it is set, otherwise fall back to any inferred configuration.
    const primaryConfiguration = configuration?.parsed ?? inferredConfiguration.parsed
    const primaryLoading = loadingRepository || loadingInferred
    const primaryError = repositoryError || inferredError

    const save = useCallback(
        async (data: SchemaCompatibleInferenceFormData) =>
            updateConfigForRepository({
                variables: { id: repoId, content: JSON.stringify(data, null, 4) },
                update: cache => cache.modify({ fields: { node: () => {} } }),
            }).then(() => {
                screenReaderAnnounce('Saved successfully')
            }),
        [updateConfigForRepository, repoId]
    )

    return (
        <div className="py-2">
            {primaryLoading ? (
                <LoadingSpinner className="d-block mx-auto mt-3" />
            ) : primaryError ? (
                <ErrorAlert error={primaryError} />
            ) : primaryConfiguration ? (
                <>
                    <InferenceForm
                        jobs={primaryConfiguration}
                        readOnly={!authenticatedUser?.siteAdmin}
                        onSubmit={data => save(data)}
                    />
                    {updatingError && <ErrorAlert error={updatingError} />}
                </>
            ) : (
                <></>
            )}
        </div>
    )
}

import React, { useCallback, useMemo, useState } from 'react'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, ErrorAlert, screenReaderAnnounce } from '@sourcegraph/wildcard'

import { useInferredConfig } from '../hooks/useInferredConfig'
import { useRepositoryConfig } from '../hooks/useRepositoryConfig'
import { useUpdateConfigurationForRepository } from '../hooks/useUpdateConfigurationForRepository'

import { autoIndexJobsToFormData } from './inference-form/auto-index-to-form-job'
import { InferenceForm } from './inference-form/InferenceForm'
import type { InferenceFormData, SchemaCompatibleInferenceFormData } from './inference-form/types'

interface ConfigurationFormProps extends TelemetryProps {
    repoId: string
    authenticatedUser: AuthenticatedUser | null
}

export const ConfigurationForm: React.FunctionComponent<ConfigurationFormProps> = ({ repoId, authenticatedUser }) => {
    const [forceInfer, setForceInfer] = useState(false)

    const { updateConfigForRepository, updatingError } = useUpdateConfigurationForRepository()
    const { inferredConfiguration, loadingInferred, inferredError } = useInferredConfig(repoId)
    const { configuration, loadingRepository, repositoryError } = useRepositoryConfig(repoId)

    const showInferButton =
        Boolean(inferredConfiguration.raw) && configuration !== null && configuration.raw !== inferredConfiguration.raw

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

    const preferredConfiguration = useMemo(() => {
        if (forceInfer) {
            return inferredConfiguration
        }

        if (configuration !== null) {
            return configuration
        }

        return inferredConfiguration
    }, [configuration, forceInfer, inferredConfiguration])

    // Show any set configuration if available, otherwise show the inferred configuration
    const preferredFormData = useMemo(
        (): InferenceFormData => autoIndexJobsToFormData({ jobs: preferredConfiguration.parsed, dirty: forceInfer }),
        [forceInfer, preferredConfiguration.parsed]
    )

    if (inferredError || repositoryError) {
        return <ErrorAlert prefix="Error fetching index configuration" error={inferredError || repositoryError} />
    }

    if (loadingInferred || loadingRepository) {
        return <LoadingSpinner className="d-block mx-auto mt-3" />
    }

    return (
        <div className="py-2">
            <InferenceForm
                key={preferredConfiguration.raw}
                initialFormData={preferredFormData}
                readOnly={!authenticatedUser?.siteAdmin}
                onSubmit={data => save(data)}
                showInferButton={showInferButton}
                onInfer={() => setForceInfer(true)}
            />
            {updatingError && <ErrorAlert error={updatingError} />}
        </div>
    )
}

import type { FC } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

import { useMutation } from '@sourcegraph/http-client'
import { Alert, Button, ErrorAlert, Link, LoadingSpinner, renderError, Tooltip } from '@sourcegraph/wildcard'

import { ExternalServiceCard } from '../../../components/externalServices/ExternalServiceCard'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import type {
    ExcludeRepoFromExternalServicesResult,
    ExcludeRepoFromExternalServicesVariables,
    SettingsAreaExternalServiceFields,
    SettingsAreaRepositoryFields,
} from '../../../graphql-operations'
import { EXCLUDE_REPO_FROM_EXTERNAL_SERVICES } from '../backend'

import { RedirectionAlert } from './RedirectionAlert'

import styles from './ExternalServiceEntry.module.scss'

interface ExternalServiceEntryProps {
    repo: SettingsAreaRepositoryFields
    service: SettingsAreaExternalServiceFields
    excludingDisabled: boolean
    /**
     * excludingLoading is true when there is a GraphQL mutation running.
     * This means that all the "exclude repo" buttons (except for the button which triggered the mutation)
     * should be disabled.
     */
    excludingLoading: boolean
    /**
     * Function to update `excludingLoading` state
     *
     * @param exclusionInProgress true if GraphQL mutation excluding the repo is in progress.
     */
    updateExclusionLoading: (exclusionInProgress: boolean) => void
    /**
     * redirectAfterExclusion is true if the redirect to code host page should happen after
     * GraphQL mutation is finished.
     */
    redirectAfterExclusion: boolean
}

export const ExternalServiceEntry: FC<ExternalServiceEntryProps> = ({
    repo,
    service,
    excludingDisabled,
    excludingLoading,
    updateExclusionLoading,
    redirectAfterExclusion,
}) => {
    const [excludeRepo, { data, error, loading: isExcluding }] = useMutation<
        ExcludeRepoFromExternalServicesResult,
        ExcludeRepoFromExternalServicesVariables
    >(EXCLUDE_REPO_FROM_EXTERNAL_SERVICES)

    return (
        <div className={styles.grid} key={service.id}>
            <div className={classNames(styles.card, data ? '' : 'mb-3')}>
                {data && !redirectAfterExclusion ? (
                    <Alert variant="success">
                        Code host configuration updated. Please see the updated code host configuration{' '}
                        <Link to={`/site-admin/external-services/${encodeURIComponent(service.id)}`}>here</Link>
                    </Alert>
                ) : (
                    <ExternalServiceCard
                        {...defaultExternalServices[service.kind]}
                        kind={service.kind}
                        title={service.displayName}
                        shortDescription="Update this code host configuration to manage repository mirroring."
                        to={`/site-admin/external-services/${encodeURIComponent(service.id)}`}
                        toIcon={null}
                        bordered={false}
                    />
                )}
                {error && <ErrorAlert error={`Failed to exclude repository: ${renderError(error)}`} />}
                {data && redirectAfterExclusion && (
                    <RedirectionAlert
                        to={`/site-admin/external-services/${encodeURIComponent(service.id)}`}
                        messagePrefix="Code host configuration updated."
                    />
                )}
            </div>
            {service.supportsRepoExclusion && !(data && !error) && (
                <div className={classNames(styles.gridButton, 'mt-3')}>
                    <Tooltip
                        content={
                            excludingDisabled
                                ? 'Excluding repository is disabled when code host configuration editing via UI is disabled'
                                : null
                        }
                    >
                        <Button
                            variant="primary"
                            className={styles.button}
                            onClick={event => {
                                event.preventDefault()
                                updateExclusionLoading(true)
                                excludeRepo({
                                    variables: {
                                        externalServices: [service.id],
                                        repo: repo.id,
                                    },
                                })
                                    .catch(
                                        // noop here is used because update error is handled directly when useMutation is called
                                        noop
                                    )
                                    .finally(() => {
                                        updateExclusionLoading(false)
                                    })
                            }}
                            disabled={excludingDisabled || (excludingLoading && !isExcluding)}
                        >
                            <span className={isExcluding ? styles.invisibleText : ''}>Exclude repository</span>
                            {isExcluding && <LoadingSpinner className={styles.loader} />}
                        </Button>
                    </Tooltip>
                </div>
            )}
        </div>
    )
}

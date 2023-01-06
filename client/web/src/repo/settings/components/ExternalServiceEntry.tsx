import { FC, useState } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useMutation } from '@sourcegraph/http-client'
import { Alert, Button, ErrorAlert, LoadingSpinner, renderError, Tooltip } from '@sourcegraph/wildcard'

import { ExternalServiceCard } from '../../../components/externalServices/ExternalServiceCard'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import {
    ExcludeRepoFromExternalServicesResult,
    ExcludeRepoFromExternalServicesVariables,
    SettingsAreaExternalServiceFields,
    SettingsAreaRepositoryFields,
} from '../../../graphql-operations'
import { EXCLUDE_REPO_FROM_EXTERNAL_SERVICES } from '../backend'

import styles from './ExternalServiceEntry.module.scss'

interface ExternalServiceEntryProps extends Pick<RouteComponentProps, 'history'> {
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
}

export const ExternalServiceEntry: FC<ExternalServiceEntryProps> = ({
    repo,
    service,
    excludingDisabled,
    excludingLoading,
    updateExclusionLoading,
    history,
}) => {
    const [ttl, setTtl] = useState<number>(3)
    const [excludeRepo, { data, error, loading: isExcluding }] = useMutation<
        ExcludeRepoFromExternalServicesResult,
        ExcludeRepoFromExternalServicesVariables
    >(EXCLUDE_REPO_FROM_EXTERNAL_SERVICES, {
        onCompleted: () => {
            let count = 3
            setInterval(() => {
                if (count === 0) {
                    history.push(`/site-admin/external-services/${service.id}`)
                }
                setTtl(count)
                count--
            }, 700)
        },
    })

    return (
        <div className={styles.grid} key={service.id}>
            <div className={classNames(styles.card, data ? '' : 'mb-3')}>
                <ExternalServiceCard
                    {...defaultExternalServices[service.kind]}
                    kind={service.kind}
                    title={service.displayName}
                    shortDescription="Update this code host configuration to manage repository mirroring."
                    to={`/site-admin/external-services/${service.id}`}
                    toIcon={null}
                    bordered={false}
                />
                {error && <ErrorAlert error={`Failed to exclude repository: ${renderError(error)}`} />}
                {data && (
                    <Alert variant="success">
                        {`Code host configuration updated. You will be redirected in ${ttl}...`}
                    </Alert>
                )}
            </div>
            {service.supportsRepoExclusion && !(data && !error) && (
                <div className={classNames(styles.button, 'mt-3')}>
                    <Tooltip
                        content={
                            excludingDisabled
                                ? 'Excluding repository is disabled when code host configuration editing via UI is disabled'
                                : null
                        }
                    >
                        <Button
                            variant="primary"
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
                            {isExcluding ? <LoadingSpinner className="mr-5 ml-5" /> : 'Exclude repository'}
                        </Button>
                    </Tooltip>
                </div>
            )}
        </div>
    )
}

import { type FC, useCallback, useEffect, useState } from 'react'

import { noop } from 'lodash'

import { useMutation, useQuery } from '@sourcegraph/http-client'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { Button, Container, ErrorAlert, H3, LoadingSpinner, renderError, Text } from '@sourcegraph/wildcard'

import { CopyableText } from '../../components/CopyableText'
import type {
    ExcludeRepoFromExternalServicesResult,
    ExcludeRepoFromExternalServicesVariables,
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    SiteExternalServiceConfigResult,
    SiteExternalServiceConfigVariables,
} from '../../graphql-operations'
import { SITE_EXTERNAL_SERVICE_CONFIG } from '../../site-admin/backend'

import { EXCLUDE_REPO_FROM_EXTERNAL_SERVICES, FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'
import { ExternalServiceEntry } from './components/ExternalServiceEntry'
import { RedirectionAlert } from './components/RedirectionAlert'

import styles from './RepoSettingsOptions.module.scss'

interface Props {
    repo: SettingsAreaRepositoryFields
}

export const RepoSettingsOptions: FC<Props> = ({ repo }) => {
    useEffect(() => {
        EVENT_LOGGER.logViewEvent('RepoSettings')
        // No need to use v2 telemetry here. This event is duplicative with 'repo.settings.mirror', 'view'
    }, [])

    const { data, error, loading } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        { variables: { name: repo.name } }
    )

    // This state shows that any of possible "exclude" buttons (in this or child components) were pushed.
    // It is used to disable all the "exclude" buttons except the button which was actually clicked.
    const [exclusionInProgress, setExclusionInProgress] = useState<boolean>(false)

    // Callback used in child components (ExternalServiceEntry) to update the state in current component.
    const updateExclusion = useCallback((updatedExclusionState: boolean) => {
        setExclusionInProgress(updatedExclusionState)
    }, [])

    const services = data?.repository?.__typename === 'Repository' && data?.repository?.externalServices.nodes

    const { data: siteConfigData, error: siteConfigError } = useQuery<
        SiteExternalServiceConfigResult,
        SiteExternalServiceConfigVariables
    >(SITE_EXTERNAL_SERVICE_CONFIG, {})

    const [excludeRepo, { data: excludeData, error: excludeError, loading: isExcluding }] = useMutation<
        ExcludeRepoFromExternalServicesResult,
        ExcludeRepoFromExternalServicesVariables
    >(EXCLUDE_REPO_FROM_EXTERNAL_SERVICES)

    const excludingDisabled =
        (!siteConfigError &&
            siteConfigData?.site?.externalServicesFromFile &&
            !siteConfigData?.site?.allowEditExternalServicesWithFile) ||
        false

    return (
        <>
            <Container className="mb-3 repo-settings-options-page">
                <H3>Repository name</H3>
                <CopyableText className="mb-3" text={repo.name} size={repo.name.length} />
                <H3>Code host connections</H3>
                {loading && <LoadingSpinner />}
                {error && <ErrorAlert error={error} />}
                {services && services.length > 0 && (
                    <div>
                        {services.map(service => (
                            <ExternalServiceEntry
                                key={service.id}
                                service={service}
                                excludingDisabled={excludingDisabled}
                                excludingLoading={exclusionInProgress}
                                updateExclusionLoading={updateExclusion}
                                repo={repo}
                                redirectAfterExclusion={services.length < 2}
                            />
                        ))}
                        {services.length > 1 && (
                            <>
                                <Text>
                                    This repository is mirrored by multiple code host connections. To change access,
                                    disable, or remove this repository, the configuration must be updated on all code
                                    host connections.
                                </Text>
                                <Button
                                    variant="primary"
                                    className={styles.button}
                                    onClick={event => {
                                        event.preventDefault()
                                        setExclusionInProgress(true)
                                        excludeRepo({
                                            variables: {
                                                externalServices: services.map(svc => svc.id),
                                                repo: repo.id,
                                            },
                                        })
                                            .catch(
                                                // noop here is used because update error is handled directly when useMutation is called
                                                noop
                                            )
                                            .finally(() => {
                                                setExclusionInProgress(false)
                                            })
                                    }}
                                    disabled={excludingDisabled || (exclusionInProgress && !isExcluding)}
                                >
                                    <span className={exclusionInProgress && isExcluding ? styles.invisibleText : ''}>
                                        Exclude repository from all code host connections
                                    </span>
                                    {exclusionInProgress && isExcluding && <LoadingSpinner className={styles.loader} />}
                                </Button>
                                {excludeError && (
                                    <ErrorAlert error={`Failed to exclude repository: ${renderError(excludeError)}`} />
                                )}
                                {excludeData && (
                                    <RedirectionAlert
                                        to="/site-admin/external-services"
                                        messagePrefix="Code host configurations updated."
                                        className="mt-2"
                                    />
                                )}
                            </>
                        )}{' '}
                    </div>
                )}
            </Container>
        </>
    )
}

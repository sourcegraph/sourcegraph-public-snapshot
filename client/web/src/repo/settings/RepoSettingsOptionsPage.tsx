import { FC, useCallback, useEffect, useState } from 'react'

import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Container, ErrorAlert, H2, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import { CopyableText } from '../../components/CopyableText'
import { PageTitle } from '../../components/PageTitle'
import {
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    SiteExternalServiceConfigResult,
    SiteExternalServiceConfigVariables,
} from '../../graphql-operations'
import { SITE_EXTERNAL_SERVICE_CONFIG } from '../../site-admin/backend'
import { eventLogger } from '../../tracking/eventLogger'

import { FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'
import { ExternalServiceEntry } from './components/ExternalServiceEntry'

interface Props extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings options page.
 */
export const RepoSettingsOptionsPage: FC<Props> = ({ repo, history }) => {
    useEffect(() => {
        eventLogger.logViewEvent('RepoSettings')
    })

    const { data, error, loading } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        { variables: { name: repo.name } }
    )

    const [exclusionInProgress, setExclusionInProgress] = useState<boolean>(false)

    const updateExclusion = useCallback((updatedExclusionState: boolean) => {
        setExclusionInProgress(updatedExclusionState)
    }, [])

    const services = data?.repository?.__typename === 'Repository' && data?.repository?.externalServices.nodes

    const { data: siteConfigData, error: siteConfigError } = useQuery<
        SiteExternalServiceConfigResult,
        SiteExternalServiceConfigVariables
    >(SITE_EXTERNAL_SERVICE_CONFIG, {})

    const excludingDisabled =
        (!siteConfigError &&
            siteConfigData?.site?.externalServicesFromFile &&
            !siteConfigData?.site?.allowEditExternalServicesWithFile) ||
        false

    return (
        <>
            <PageTitle title="Repository settings" />
            <PageHeader path={[{ text: 'Settings' }]} headingElement="h2" className="mb-3" />
            <Container className="repo-settings-options-page">
                <H2 className="mb-3">Repository name</H2>
                <CopyableText className="mb-3" text={repo.name} size={repo.name.length} />
                <H2 className="mb-3">Code hosts</H2>
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
                                history={history}
                            />
                        ))}
                        {services.length > 1 && (
                            <>
                                <Text className="text-muted">
                                    This repository is mirrored from multiple code hosts. To remove the repository from
                                    this Sourcegraph instance, all code host configurations need to be updated.
                                </Text>
                            </>
                        )}
                    </div>
                )}
            </Container>
        </>
    )
}

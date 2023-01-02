import { FC, useEffect } from 'react'

import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import { Container, ErrorAlert, H2, LoadingSpinner, PageHeader, Text } from '@sourcegraph/wildcard'

import { CopyableText } from '../../components/CopyableText'
import { ExternalServiceCard } from '../../components/externalServices/ExternalServiceCard'
import { defaultExternalServices } from '../../components/externalServices/externalServices'
import { PageTitle } from '../../components/PageTitle'
import {
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { FETCH_SETTINGS_AREA_REPOSITORY_GQL } from './backend'

interface Props extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings options page.
 */
export const RepoSettingsOptionsPage: FC<Props> = ({ repo: propsRepo }) => {
    useEffect(() => {
        eventLogger.logViewEvent('RepoSettings')
    })

    const { data, error, loading } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        { variables: { name: propsRepo.name } }
    )

    const services = data?.repository?.__typename === 'Repository' && data?.repository?.externalServices.nodes

    return (
        <>
            <PageTitle title="Repository settings" />
            <PageHeader path={[{ text: 'Settings' }]} headingElement="h2" className="mb-3" />
            <Container className="repo-settings-options-page">
                <H2 className="mb-3">Code hosts</H2>
                {loading && <LoadingSpinner />}
                {error && <ErrorAlert error={error} />}
                {services && services.length > 0 && (
                    <div className="mb-3">
                        {services.map(service => (
                            <div className="mb-3" key={service.id}>
                                <ExternalServiceCard
                                    {...defaultExternalServices[service.kind]}
                                    kind={service.kind}
                                    title={service.displayName}
                                    shortDescription="Update this external service configuration to manage repository mirroring."
                                    to={`/site-admin/external-services/${service.id}`}
                                    toIcon={null}
                                    bordered={false}
                                />
                            </div>
                        ))}
                        {services.length > 1 && (
                            <Text>
                                This repository is mirrored by multiple external services. To change access, disable, or
                                remove this repository, the configuration must be updated on all external services.
                            </Text>
                        )}
                    </div>
                )}
                <H2 className="mb-3">Repository name</H2>
                <CopyableText text={propsRepo.name} size={propsRepo.name.length} />
            </Container>
        </>
    )
}

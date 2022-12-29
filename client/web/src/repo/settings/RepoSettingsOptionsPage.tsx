import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { asError } from '@sourcegraph/common'
import { Container, PageHeader, LoadingSpinner, Text, ErrorAlert, H2 } from '@sourcegraph/wildcard'

import { CopyableText } from '../../components/CopyableText'
import { ExternalServiceCard } from '../../components/externalServices/ExternalServiceCard'
import { defaultExternalServices } from '../../components/externalServices/externalServices'
import { PageTitle } from '../../components/PageTitle'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { fetchSettingsAreaRepository } from './backend'
import { FC, useEffect, useState } from 'react'

interface Props extends RouteComponentProps<{}> {
    repo: SettingsAreaRepositoryFields
}

/**
 * The repository settings options page.
 */
export const RepoSettingsOptionsPage: FC<Props> = ({ repo: propsRepo }) => {
    const repoUpdates = new Subject<void>()
    const subscriptions = new Subscription()

    const [loading] = useState<boolean>(false)
    const [repository, setRepository] = useState<SettingsAreaRepositoryFields>(propsRepo)
    const [error, setError] = useState<string | undefined>(undefined)

    useEffect(() => {
        eventLogger.logViewEvent('RepoSettings')
        subscriptions.add(
            repoUpdates.pipe(switchMap(() => fetchSettingsAreaRepository(propsRepo.name))).subscribe(
                repoLocal => setRepository(repoLocal),
                errorLocal => setError(asError(errorLocal).message)
            )
        )
        return function cleanup() {
            subscriptions.unsubscribe()
        }
    })

    const services = repository.externalServices.nodes

    return (
        <>
            <PageTitle title="Repository settings" />
            <PageHeader path={[{ text: 'Settings' }]} headingElement="h2" className="mb-3" />
            <Container className="repo-settings-options-page">
                <H2 className="mb-3">Code hosts</H2>
                {loading && <LoadingSpinner />}
                {error && <ErrorAlert error={error} />}
                {services.length > 0 && (
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
                                This repository is mirrored by multiple external services. To change access,
                                disable, or remove this repository, the configuration must be updated on all
                                external services.
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

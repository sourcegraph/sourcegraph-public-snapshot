import { useEffect, useMemo } from 'react'

import { mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Badge, Container, ErrorAlert, H3, Icon, Link, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { GlobalCodeIntelStatusResult } from '../../../../graphql-operations'
import { ExternalRepositoryIcon } from '../../../../site-admin/components/ExternalRepositoryIcon'
import { globalCodeIntelStatusQuery } from '../backend'

import { DataSummary, DataSummaryItem } from '../components/DataSummary'

import styles from './GlobalDashboardPage.module.scss'

interface GlobalDashboardPageProps extends TelemetryProps {}

export const GlobalDashboardPage: React.FunctionComponent<GlobalDashboardPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logPageView('CodeIntelGlobalDashboard')
    }, [telemetryService])

    const { data, error, loading } = useQuery<GlobalCodeIntelStatusResult>(globalCodeIntelStatusQuery, {
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
    })

    const summaryItems = useMemo((): DataSummaryItem[] => {
        if (!data) {
            return []
        }

        return [
            {
                label: 'Repositories with precise code intelligence',
                value: data.codeIntelSummary.numRepositoriesWithCodeIntelligence,
                className: styles.summaryItemExtended,
                valueClassName: 'text-success',
            },
            {
                label: 'Repositories with errors',
                value: data.codeIntelSummary.repositoriesWithErrors?.nodes.length || 0,
                valueClassName: 'text-danger',
            },
            {
                label: 'Configurable repositories',
                value: data.codeIntelSummary.repositoriesWithConfiguration?.nodes.length || 0,
                valueClassName: 'text-merged',
            },
        ]
    }, [data])

    if (loading || !data) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert prefix="Failed to load code intelligence summary" error={error} />
    }

    return (
        <>
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Code intelligence summary</>,
                    },
                ]}
                className="mb-3"
            />
            <Container>
                <DataSummary items={summaryItems} className="pb-3" />

                {data.codeIntelSummary.repositoriesWithErrors &&
                    data.codeIntelSummary.repositoriesWithErrors.nodes.length > 0 && (
                        <div className={styles.details}>
                            <H3 className="px-3">Repositories with errors</H3>

                            <ul className={styles.detailsList}>
                                {data.codeIntelSummary.repositoriesWithErrors.nodes.map(({ repository, count }) => (
                                    <li key={repository.name} className={styles.detailsListItem}>
                                        <div>
                                            {repository.externalRepository && (
                                                <ExternalRepositoryIcon
                                                    externalRepo={{
                                                        serviceID: repository.externalRepository.serviceID,
                                                        serviceType: repository.externalRepository.serviceType,
                                                    }}
                                                />
                                            )}
                                            <RepoLink
                                                repoName={repository.name}
                                                to={`${repository.url}/-/code-graph/dashboard`}
                                            />
                                        </div>
                                        <Link
                                            to={`${repository.url}/-/code-graph/dashboard`} // TODO: Link to list of errors for repo specific
                                            className={styles.detailsLink}
                                        >
                                            <Badge variant="danger" className={styles.badge} pill={true}>
                                                {count} {count > 1 ? 'errors' : 'error'}
                                            </Badge>
                                            <Icon svgPath={mdiChevronRight} size="md" aria-label="Fix" />
                                        </Link>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}

                {data.codeIntelSummary.repositoriesWithConfiguration &&
                    data.codeIntelSummary.repositoriesWithConfiguration.nodes.length > 0 && (
                        <div className={styles.details}>
                            <H3 className="px-3">Configurable repositories</H3>

                            <ul className={styles.detailsList}>
                                {data.codeIntelSummary.repositoriesWithConfiguration.nodes.map(
                                    ({ repository, indexers }) => (
                                        <li key={repository.name} className={styles.detailsListItem}>
                                            <div>
                                                {repository.externalRepository && (
                                                    <ExternalRepositoryIcon
                                                        externalRepo={{
                                                            serviceID: repository.externalRepository.serviceID,
                                                            serviceType: repository.externalRepository.serviceType,
                                                        }}
                                                    />
                                                )}
                                                <RepoLink
                                                    repoName={repository.name}
                                                    to={`${repository.url}/-/code-graph/dashboard`}
                                                />
                                            </div>
                                            <Link
                                                to={`${repository.url}/-/code-graph/dashboard`} // TODO: Link to list of actions for repo specific
                                                className={styles.detailsLink}
                                            >
                                                <Badge variant="info" className={styles.badge} pill={true}>
                                                    {indexers.length} {indexers.length > 1 ? 'actions' : 'action'}
                                                </Badge>
                                                <Icon svgPath={mdiChevronRight} size="md" aria-label="Configure" />
                                            </Link>
                                        </li>
                                    )
                                )}
                            </ul>
                        </div>
                    )}
            </Container>
        </>
    )
}

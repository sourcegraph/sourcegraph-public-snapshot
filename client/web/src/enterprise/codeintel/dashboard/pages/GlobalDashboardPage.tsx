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

        const countWithPreciseCodeIntel = data.codeIntelSummary.numRepositoriesWithCodeIntelligence
        const countWithErrors = data.codeIntelSummary.repositoriesWithErrors?.nodes.length || 0
        const countConfigurable = data.codeIntelSummary.repositoriesWithConfiguration?.nodes.length || 0

        return [
            {
                label: `${
                    countWithPreciseCodeIntel === 1 ? 'Repository' : 'Repositories'
                } with precise code intelligence`,
                value: countWithPreciseCodeIntel,
                className: styles.summaryItemExtended,
                valueClassName: 'text-success',
            },
            {
                label: `${countWithErrors === 1 ? 'Repository' : 'Repositories'} with errors`,
                value: countWithErrors,
                valueClassName: 'text-danger',
            },
            {
                label: `Configurable ${countConfigurable === 1 ? 'repository' : 'repositories'}`,
                value: countConfigurable,
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
                {/* TODO: Make data summary links to configure? */}
                <DataSummary items={summaryItems} className="pb-3" />
            </Container>

            <Container className="mt-3">
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
                                                to={`${repository.url}/-/code-graph/dashboard?show=errors`}
                                            />
                                        </div>
                                        <Link
                                            to={`${repository.url}/-/code-graph/dashboard?show=errors`}
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
                                                    to={`${repository.url}/-/code-graph/dashboard?show=suggestions`}
                                                />
                                            </div>
                                            <Link
                                                to={`${repository.url}/-/code-graph/dashboard?show=suggestions`}
                                                className={styles.detailsLink}
                                            >
                                                <Badge variant="info" className={styles.badge} pill={true}>
                                                    {indexers.length}{' '}
                                                    {indexers.length > 1 ? 'suggestions' : 'suggestion'}
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

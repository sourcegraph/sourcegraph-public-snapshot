import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ArrowLeftBoldIcon from 'mdi-react/ArrowLeftBoldIcon'
import ArrowRightBoldIcon from 'mdi-react/ArrowRightBoldIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import { RouteComponentProps } from 'react-router'
import { Observable, of, timer } from 'rxjs'
import { catchError, concatMap, delay, map, repeatWhen, takeWhile } from 'rxjs/operators'
import { parse as _parseVersion, SemVer } from 'semver'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable, Alert, Icon, Typography } from '@sourcegraph/wildcard'

import { Collapsible } from '../components/Collapsible'
import { FilteredConnection, FilteredConnectionFilter, Connection } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { OutOfBandMigrationFields } from '../graphql-operations'

import {
    fetchAllOutOfBandMigrations as defaultFetchAllMigrations,
    fetchSiteUpdateCheck as defaultFetchSiteUpdateCheck,
} from './backend'

import styles from './SiteAdminMigrationsPage.module.scss'

export interface SiteAdminMigrationsPageProps extends RouteComponentProps<{}>, TelemetryProps {
    fetchAllMigrations?: typeof defaultFetchAllMigrations
    fetchSiteUpdateCheck?: () => Observable<{ productVersion: string }>
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Migration state',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all migrations',
                args: {},
            },
            {
                label: 'Pending',
                value: 'pending',
                tooltip: 'Show pending migrations',
                args: { completed: false },
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed migrations',
                args: { completed: true },
            },
        ],
    },
]

/* How frequently to refresh data from the API. */
const REFRESH_INTERVAL_MS = 5000

/* How many (minor) versions we can upgrade at once. */
const UPGRADE_RANGE = 1

/* How many (minor) versions we can downgrade at once. */
const DOWNGRADE_RANGE = 1

export const SiteAdminMigrationsPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminMigrationsPageProps>
> = ({
    fetchAllMigrations = defaultFetchAllMigrations,
    fetchSiteUpdateCheck = defaultFetchSiteUpdateCheck,
    now,
    telemetryService,
    ...props
}) => {
    const migrationsOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, undefined).pipe(
                    concatMap(() =>
                        fetchAllMigrations().pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            repeatWhen(observable => observable.pipe(delay(REFRESH_INTERVAL_MS)))
                        )
                    ),
                    takeWhile(() => true, true)
                ),
            [fetchAllMigrations]
        )
    )

    const queryMigrations = useCallback(
        ({
            query,
            completed,
        }: {
            query?: string
            completed?: boolean
        }): Observable<Connection<OutOfBandMigrationFields>> => {
            if (isErrorLike(migrationsOrError) || migrationsOrError === undefined) {
                return of({ nodes: [] })
            }

            return of({
                nodes: migrationsOrError.filter(
                    migration =>
                        (completed === undefined || completed === isComplete(migration)) &&
                        (!query || matchesQuery(migration, query))
                ),
                totalCount: migrationsOrError.length,
                pageInfo: { hasNextPage: false },
            })
        },
        [migrationsOrError]
    )

    return (
        <div className="site-admin-migrations-page">
            {isErrorLike(migrationsOrError) ? (
                <ErrorAlert prefix="Error loading out of band migrations" error={migrationsOrError} />
            ) : migrationsOrError === undefined ? (
                <LoadingSpinner />
            ) : (
                <>
                    <PageTitle title="Out of band migrations - Admin" />
                    <Typography.H2>Out-of-band migrations</Typography.H2>

                    <p>
                        Out-of-band migrations run in the background of the Sourcegraph instance convert data from an
                        old format into a new format. Consult this page prior to upgrading your Sourcegraph instance to
                        ensure that all expected migrations have completed.
                    </p>

                    <MigrationBanners migrations={migrationsOrError} fetchSiteUpdateCheck={fetchSiteUpdateCheck} />

                    <div className="list-group">
                        <FilteredConnection<OutOfBandMigrationFields, Omit<MigrationNodeProps, 'node'>>
                            listComponent="div"
                            listClassName={classNames('mb-3', styles.migrationsGrid)}
                            noun="migration"
                            pluralNoun="migrations"
                            queryConnection={queryMigrations}
                            nodeComponent={MigrationNode}
                            nodeComponentProps={{ now }}
                            history={props.history}
                            location={props.location}
                            filters={filters}
                        />
                    </div>
                </>
            )}
        </div>
    )
}

interface MigrationBannersProps {
    migrations: OutOfBandMigrationFields[]
    fetchSiteUpdateCheck?: () => Observable<{ productVersion: string }>
}

const MigrationBanners: React.FunctionComponent<React.PropsWithChildren<MigrationBannersProps>> = ({
    migrations,
    fetchSiteUpdateCheck = defaultFetchSiteUpdateCheck,
}) => {
    const productVersion = useObservable(
        useMemo(() => fetchSiteUpdateCheck().pipe(map(site => parseVersion(site.productVersion))), [
            fetchSiteUpdateCheck,
        ])
    )
    if (!productVersion) {
        return <></>
    }

    const nextVersion = parseVersion(`${productVersion.major}.${productVersion.minor + UPGRADE_RANGE}.0`)
    const previousVersion = parseVersion(`${productVersion.major}.${productVersion.minor - DOWNGRADE_RANGE}.0`)

    const invalidMigrations = migrationsInvalidForVersion(migrations, productVersion)
    const invalidMigrationsAfterUpgrade = migrationsInvalidForVersion(migrations, nextVersion)
    const invalidMigrationsAfterDowngrade = migrationsInvalidForVersion(migrations, previousVersion)

    if (invalidMigrations.length > 0) {
        return <MigrationInvalidBanner migrations={invalidMigrations} />
    }
    return (
        <>
            {invalidMigrationsAfterUpgrade.length > 0 && (
                <MigrationUpgradeWarningBanner migrations={invalidMigrationsAfterUpgrade} />
            )}
            {invalidMigrationsAfterDowngrade.length > 0 && (
                <MigrationDowngradeWarningBanner migrations={invalidMigrationsAfterDowngrade} />
            )}
        </>
    )
}

interface MigrationInvalidBannerProps {
    migrations: OutOfBandMigrationFields[]
}

const MigrationInvalidBanner: React.FunctionComponent<React.PropsWithChildren<MigrationInvalidBannerProps>> = ({
    migrations,
}) => (
    <Alert variant="danger">
        <p>
            <Icon className="mr-2" as={AlertCircleIcon} />
            <strong>Contact support.</strong> The following migrations are not in the expected state. You have partially
            migrated or un-migrated data in a format that is incompatible with the currently deployed version of
            Sourcegraph.{' '}
            <strong>Continuing to run your instance in this state will result in errors and possible data loss.</strong>
        </p>

        <ul className="mb-0">
            {migrations.map(migration => (
                <li key={migration.id}>{migration.description}</li>
            ))}
        </ul>
    </Alert>
)

interface MigrationUpgradeWarningBannerProps {
    migrations: OutOfBandMigrationFields[]
}

const MigrationUpgradeWarningBanner: React.FunctionComponent<
    React.PropsWithChildren<MigrationUpgradeWarningBannerProps>
> = ({ migrations }) => (
    <Alert variant="warning">
        <p>
            The next version of Sourcegraph removes support for reading an old data format. Your Sourcegraph instance
            must complete the following migrations to ensure your data remains readable.{' '}
            <strong>If you upgrade your Sourcegraph instance now, you may corrupt or lose data.</strong>
        </p>
        <ul>
            {migrations.map(migration => (
                <li key={migration.id}>{migration.description}</li>
            ))}
        </ul>
        <span>Contact support if these migrations are not making progress or if there are associated errors.</span>
    </Alert>
)

interface MigrationDowngradeWarningBannerProps {
    migrations: OutOfBandMigrationFields[]
}

const MigrationDowngradeWarningBanner: React.FunctionComponent<
    React.PropsWithChildren<MigrationDowngradeWarningBannerProps>
> = ({ migrations }) => (
    <Alert variant="warning">
        <p>
            <Icon className="mr-2" as={WarningIcon} />
            <span>
                The previous version of Sourcegraph does not support reading data that has been migrated into a new
                format. Your Sourcegraph instance must undo the following migrations to ensure your data can be read by
                the previous version.{' '}
                <strong>If you downgrade your Sourcegraph instance now, you may corrupt or lose data.</strong>
            </span>
        </p>

        <ul>
            {migrations.map(migration => (
                <li key={migration.id}>{migration.description}</li>
            ))}
        </ul>

        <span>Contact support for assistance with downgrading your instance.</span>
    </Alert>
)

interface MigrationNodeProps {
    node: OutOfBandMigrationFields
    now?: () => Date
}

const MigrationNode: React.FunctionComponent<React.PropsWithChildren<MigrationNodeProps>> = ({ node, now }) => (
    <React.Fragment key={node.id}>
        <span className={styles.separator} />

        <div className={classNames('d-flex flex-column', styles.information)}>
            <div>
                <Typography.H3>{node.description}</Typography.H3>

                <p className="m-0">
                    <span className="text-muted">Team</span> <strong>{node.team}</strong>{' '}
                    <span className="text-muted">is migrating data in</span> <strong>{node.component}</strong>
                    <span className="text-muted">.</span>
                </p>

                <p className="m-0">
                    <span className="text-muted">Began running in v</span>
                    {node.introduced}
                    {node.deprecated && (
                        <>
                            {' '}
                            <span className="text-muted">and will cease running in v</span>
                            {node.deprecated}
                        </>
                    )}
                    .
                </p>
            </div>
        </div>

        <span className={classNames('d-none d-md-inline', styles.progress)}>
            <div className="m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
                <div>
                    {node.applyReverse ? (
                        <Icon className="mr-1 text-danger" as={ArrowLeftBoldIcon} />
                    ) : (
                        <Icon className="mr-1" as={ArrowRightBoldIcon} />
                    )}
                    {Math.floor(node.progress * 100)}%
                </div>

                <div>
                    <meter
                        min={0}
                        low={0.2}
                        high={0.8}
                        max={1}
                        optimum={1}
                        value={node.progress}
                        data-tooltip={`${Math.floor(node.progress * 100)}%`}
                        aria-label="migration progress"
                        data-placement="bottom"
                    />
                </div>

                {node.lastUpdated && node.lastUpdated !== '' && (
                    <>
                        <div className="text-center">
                            <span className="text-muted">Last updated</span>
                        </div>
                        <div className="text-center">
                            <small>
                                <Timestamp date={node.lastUpdated} now={now} noAbout={true} />
                            </small>
                        </div>
                    </>
                )}
            </div>
        </span>

        {node.errors.length > 0 && (
            <Collapsible
                title={<strong>Recent errors ({node.errors.length})</strong>}
                className="p-0 font-weight-normal"
                buttonClassName="mb-0"
                titleAtStart={true}
                defaultExpanded={false}
            >
                <div className={classNames('pt-2', styles.nodeGrid)}>
                    {node.errors
                        .map((error, index) => ({ ...error, index }))
                        .map(error => (
                            <React.Fragment key={error.index}>
                                <div className="py-1 pr-2">
                                    <Timestamp date={error.created} now={now} />
                                </div>

                                <span className={classNames('py-1 pl-2', styles.nodeGridCode)}>
                                    <code>{error.message}</code>
                                </span>
                            </React.Fragment>
                        ))}
                </div>
            </Collapsible>
        )}
    </React.Fragment>
)

type PartialVersion = SemVer | null

/** Parse the given version safely. */
const parseVersion = (version: string): PartialVersion => {
    try {
        return _parseVersion(version)
    } catch {
        return null
    }
}

/** Returns true if the given migration state is invalid for the given version. */
export const isInvalidForVersion = (migration: OutOfBandMigrationFields, version: PartialVersion): boolean => {
    if (!version) {
        return false
    }

    // Migrations only store major/minor version components
    const introduced = parseVersion(`${migration.introduced}.0`)
    if (introduced && version.major === introduced.major && version.minor < introduced.minor) {
        return migration.progress !== 0 && !migration.nonDestructive
    }

    if (migration.deprecated) {
        // Migrations only store major/minor version components
        const deprecated = parseVersion(`${migration.deprecated}.0`)
        if (deprecated && version.major === deprecated.major && version.minor >= deprecated.minor) {
            return migration.progress !== 1
        }
    }

    return false
}

/** Returns the set of migrations that are invalid for the given version. */
const migrationsInvalidForVersion = (
    migrations: OutOfBandMigrationFields[],
    version: PartialVersion
): OutOfBandMigrationFields[] => migrations.filter(migration => isInvalidForVersion(migration, version))

/** Returns true if the given migration is has completed (100% if forward, 0% if reverse). */
export const isComplete = (migration: OutOfBandMigrationFields): boolean =>
    (migration.progress === 0 && migration.applyReverse) || (migration.progress === 1 && !migration.applyReverse)

/** Returns the searchable text from a migration. */
const searchFields = (migration: OutOfBandMigrationFields): string[] => [
    migration.team,
    migration.component,
    migration.description,
    migration.introduced,
    migration.deprecated || '',
    ...migration.errors.map(error => error.message),
]

/** Returns true if the migration matches the given query. */
const matchesQuery = (migration: OutOfBandMigrationFields, query: string): boolean => {
    const fields = searchFields(migration)
        .map(value => value.toLowerCase())
        .filter(value => value !== '')

    return query
        .toLowerCase()
        .split(' ')
        .filter(query => query !== '')
        .every(query => fields.some(value => value.includes(query)))
}

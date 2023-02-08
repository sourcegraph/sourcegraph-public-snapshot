import { useEffect, useMemo, useState } from 'react'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, ErrorAlert, Link, LoadingSpinner, PageHeader, Input, useDebounce } from '@sourcegraph/wildcard'

import {
    buildFilterArgs,
    FilterControl,
    FilteredConnectionFilter,
    FilteredConnectionFilterValue,
} from '../components/FilteredConnection'
import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import { getFilterFromURL } from '../components/FilteredConnection/utils'
import { PageTitle } from '../components/PageTitle'
import { PackagesResult, PackagesVariables, SiteAdminPackageFields } from '../graphql-operations'

import { PACKAGES_QUERY } from './backend'
import { PackageHost, PackageRepositoryIcon } from './components/PackageRepositoryIcon'

interface SiteAdminPackagesPageProps extends TelemetryProps {}

interface PackageNodeProps {
    node: SiteAdminPackageFields
}

const PackageNode: React.FunctionComponent<React.PropsWithChildren<PackageNodeProps>> = ({ node }) => (
    <li className="list-group-item px-0 py-2">
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <PackageRepositoryIcon host={node.scheme as PackageHost} />
                {node.repository ? <RepoLink repoName={node.name} to={node.repository.url} /> : node.name}
                {/* <RepoMirrorInfo mirrorInfo={node.mirrorInfo} /> */}
            </div>
        </div>
    </li>
)

const PACKAGE_HOST_FILTER: Record<PackageHost, FilteredConnectionFilterValue> = {
    npm: {
        label: 'npm',
        value: 'npm',
        args: {
            scheme: 'npm',
        },
    },
    go: {
        label: 'Go',
        value: 'go',
        args: {
            scheme: 'go',
        },
    },
    semanticdb: {
        label: 'SemanticDB',
        value: 'semanticdb',
        args: {
            scheme: 'semanticdb',
        },
    },
    'scip-ruby': {
        label: 'Ruby',
        value: 'scip-ruby',
        args: {
            scheme: 'scip-ruby',
        },
    },
    python: {
        label: 'Python',
        value: 'python',
        args: {
            scheme: 'python',
        },
    },
    'rust-analyzer': {
        label: 'Rust',
        value: 'rust-analyzer',
        args: {
            scheme: 'rust-analyzer',
        },
    },
}

const FILTERS: FilteredConnectionFilter[] = [
    {
        id: 'host',
        label: 'Code Host',
        type: 'select',
        values: Object.values(PACKAGE_HOST_FILTER),
    },
]

/**
 * A page displaying the packages on this instance.
 */
export const SiteAdminPackagesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminPackagesPageProps>> = ({
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminPackages')
    }, [telemetryService])

    const [filterValues, setFilterValues] = useState<Map<string, FilteredConnectionFilterValue>>(() =>
        getFilterFromURL(new URLSearchParams(location.search), FILTERS)
    )

    const [searchValue, setSearchValue] = useState<string>(
        () => new URLSearchParams(location.search).get('query') || ''
    )

    const query = useDebounce(searchValue, 200)

    const variables = useMemo(() => {
        const args = buildFilterArgs(filterValues)

        return {
            ...args,
            name: query,
            first: 10,
            after: null,
        }
    }, [filterValues, query])

    const { connection, error, loading } = useShowMorePagination<
        PackagesResult,
        PackagesVariables,
        SiteAdminPackageFields
    >({
        query: PACKAGES_QUERY,
        variables,
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            return data.packageRepoReferences
        },
    })

    return (
        <div>
            <PageTitle title="Packages - Admin" />
            <PageHeader
                path={[{ text: 'Packages' }]}
                headingElement="h2"
                description={
                    <>
                        Packages are synced from connected{' '}
                        <Link
                            to="/site-admin/external-services"
                            data-testid="test-repositories-code-host-connections-link"
                        >
                            code hosts
                        </Link>
                        .
                    </>
                }
                className="mb-3"
            />

            <Container className="mb-3">
                <div className="d-flex justify-content-center">
                    <FilterControl
                        filters={FILTERS}
                        values={filterValues}
                        onValueSelect={(filter: FilteredConnectionFilter, value: FilteredConnectionFilterValue) =>
                            setFilterValues(values => {
                                const newValues = new Map(values)
                                newValues.set(filter.id, value)
                                return newValues
                            })
                        }
                    />
                    <Input
                        type="search"
                        className="flex-1"
                        placeholder="Search packages..."
                        name="query"
                        value={searchValue}
                        onChange={event => setSearchValue(event.currentTarget.value)}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                        aria-label="Search packages..."
                        variant="regular"
                    />
                </div>
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner className="d-block mx-auto mt-3" />}
                <ul className="list-group list-group-flush mt-4">
                    {(connection?.nodes || []).map(node => (
                        <PackageNode node={node} key={node.id} />
                    ))}
                </ul>
            </Container>
        </div>
    )
}

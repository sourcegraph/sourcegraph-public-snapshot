import { useEffect, useMemo, useState } from 'react'

import { isEqual } from 'lodash'
import { useNavigate, useLocation } from 'react-router-dom-v5-compat'

import { dataOrThrowErrors, useQuery } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    ErrorAlert,
    Link,
    LoadingSpinner,
    PageHeader,
    Input,
    useDebounce,
    Button,
} from '@sourcegraph/wildcard'

import {
    buildFilterArgs,
    FilterControl,
    FilteredConnectionFilter,
    FilteredConnectionFilterValue,
} from '../../components/FilteredConnection'
import { useShowMorePagination } from '../../components/FilteredConnection/hooks/useShowMorePagination'
import { getFilterFromURL, getUrlQuery } from '../../components/FilteredConnection/utils'
import { PageTitle } from '../../components/PageTitle'
import {
    PackagesResult,
    PackagesVariables,
    SiteAdminPackageFields,
    ExternalServiceKindsVariables,
    ExternalServiceKindsResult,
    ExternalServiceKind,
} from '../../graphql-operations'
import { EXTERNAL_SERVICE_KINDS, PACKAGES_QUERY } from '../backend'
import { PackageHost, PackageRepositoryIcon } from '../components/PackageRepositoryIcon'
import { RepoMirrorInfo } from '../components/RepoMirrorInfo'
import { PackageRepoReferenceKind } from '@sourcegraph/shared/src/graphql-operations'

import { mdiCloudQuestion } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../components/externalServices/externalServices'
import { ExternalServiceKind } from '../../graphql-operations'

interface SiteAdminPackagesPageProps extends TelemetryProps {}

interface PackageRepositoryIconProps {
    host: PackageRepoReferenceKind
}

export const PackageRepositoryIcon: React.FunctionComponent<PackageRepositoryIconProps> = ({ host }) => {
    const IconComponent = defaultExternalServices[PACKAGE_HOST_TO_EXTERNAL_REPO[host]].icon
    return IconComponent ? (
        <Icon as={IconComponent} aria-label="Package host logo" className="mr-2" />
    ) : (
        <Icon svgPath={mdiCloudQuestion} aria-label="Unknown package host" className="mr-2" />
    )
}

interface PackageNodeProps {
    node: SiteAdminPackageFields
}

const PackageNode: React.FunctionComponent<React.PropsWithChildren<PackageNodeProps>> = ({ node }) => (
    <li className="list-group-item px-0 py-2">
        <div className="d-flex align-items-center justify-content-between">
            <div>
                <PackageRepositoryIcon host={node.scheme} />
                {node.repository ? <RepoLink repoName={node.name} to={node.repository.url} /> : node.name}
                {node.repository && <RepoMirrorInfo mirrorInfo={node.repository.mirrorInfo} />}
            </div>
        </div>
    </li>
)

interface FriendlyPackageRepoReferenceKind {
    label: string
    value: PackageRepoReferenceKind
}

const EXTERNAL_SERVICE_KIND_TO_PACKAGE_REPO_REFERENCE_KIND: Partial<
    Record<ExternalServiceKind, FriendlyPackageRepoReferenceKind>
> = {
    [ExternalServiceKind.NPMPACKAGES]: {
        label: 'NPM',
        value: PackageRepoReferenceKind.NPMPACKAGES,
    },
    [ExternalServiceKind.GOMODULES]: {
        label: 'Go',
        value: PackageRepoReferenceKind.GOMODULES,
    },
    [ExternalServiceKind.JVMPACKAGES]: {
        label: 'JVM',
        value: PackageRepoReferenceKind.JVMPACKAGES,
    },
    [ExternalServiceKind.RUBYPACKAGES]: {
        label: 'Ruby',
        value: PackageRepoReferenceKind.RUBYPACKAGES,
    },
    [ExternalServiceKind.PYTHONPACKAGES]: {
        label: 'Python',
        value: PackageRepoReferenceKind.PYTHONPACKAGES,
    },
    [ExternalServiceKind.RUSTPACKAGES]: {
        label: 'Rust',
        value: PackageRepoReferenceKind.RUSTPACKAGES,
    },
}

/**
 * A page displaying the packages on this instance.
 */
export const SiteAdminPackagesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminPackagesPageProps>> = ({
    telemetryService,
}) => {
    const location = useLocation()
    const navigate = useNavigate()

    useEffect(() => {
        telemetryService.logPageView('SiteAdminPackages')
    }, [telemetryService])

    const {
        loading: extSvcLoading,
        data: extSvcs,
        error: extSvcError,
    } = useQuery<ExternalServiceKindsResult, ExternalServiceKindsVariables>(EXTERNAL_SERVICE_KINDS, {})

    const filters = useMemo<FilteredConnectionFilter[]>(() => {
        const values = [
            {
                label: 'All',
                value: 'all',
                args: {},
            },
        ]

        for (const extSvc of extSvcs?.externalServices.nodes ?? []) {
            const packageRepoScheme = EXTERNAL_SERVICE_KIND_TO_PACKAGE_REPO_REFERENCE_KIND[extSvc.kind]

            if (packageRepoScheme) {
                values.push({
                    ...packageRepoScheme,
                    args: { scheme: packageRepoScheme.value },
                })
            }
        }

        return [
            {
                id: 'ecosystem',
                label: 'Host',
                type: 'select',
                values,
            },
        ]
    }, [extSvcs])

    const [filterValues, setFilterValues] = useState<Map<string, FilteredConnectionFilterValue>>(() =>
        getFilterFromURL(new URLSearchParams(location.search), filters)
    )

    const [searchValue, setSearchValue] = useState<string>(
        () => new URLSearchParams(location.search).get('query') || ''
    )

    const query = useDebounce(searchValue, 200)

    useEffect(() => {
        const searchFragment = getUrlQuery({
            query: searchValue,
            filters,
            filterValues,
            search: location.search,
        })
        const searchFragmentParams = new URLSearchParams(searchFragment)
        searchFragmentParams.sort()

        const oldParams = new URLSearchParams(location.search)
        oldParams.sort()

        if (!isEqual(Array.from(searchFragmentParams), Array.from(oldParams))) {
            navigate(
                {
                    search: searchFragment,
                    hash: location.hash,
                },
                {
                    replace: true,
                    // Do not throw away flash messages
                    state: location.state,
                }
            )
        }
    }, [filterValues, filters, searchValue, location, navigate])

    const variables = useMemo<PackagesVariables>(() => {
        const args = buildFilterArgs(filterValues)

        return {
            name: query,
            scheme: null,
            after: null,
            first: 15,
            ...args,
        }
    }, [filterValues, query])

    const { connection, error, loading, fetchMore, hasNextPage } = useShowMorePagination<
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
        options: {
            fetchPolicy: 'cache-first',
            useURL: true,
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
                <div className="d-flex mt-3">
                    <FilterControl
                        filters={filters}
                        values={filterValues}
                        onValueSelect={(filter: FilteredConnectionFilter, value: FilteredConnectionFilterValue) =>
                            setFilterValues(values => {
                                const newValues = new Map(values)
                                newValues.set(filter.id, value)
                                return newValues
                            })
                        }
                    />
                </div>
                {error && !loading && <ErrorAlert error={error} />}
                {loading && !error && <LoadingSpinner className="d-block mx-auto mt-3" />}
                {connection?.nodes && connection.nodes.length > 0 && (
                    <ul className="list-group list-group-flush mt-4">
                        {(connection?.nodes || []).map(node => (
                            <PackageNode node={node} key={node.id} />
                        ))}
                    </ul>
                )}

                {hasNextPage && (
                    <div>
                        <Button variant="link" size="sm" onClick={fetchMore}>
                            Show more
                        </Button>
                    </div>
                )}
            </Container>
        </div>
    )
}

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
    Alert,
    Text,
    Code,
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
import { ExternalRepositoryIcon } from '../components/ExternalRepositoryIcon'
import { RepoMirrorInfo } from '../components/RepoMirrorInfo'

import styles from './SiteAdminPackagesPage.module.scss'

// TODO: Share this with backend (Make `scheme` a GQL enum)
type PackageScheme = 'npm' | 'go' | 'semanticdb' | 'scip-ruby' | 'python' | 'rust-analyzer'

interface PackageNodeProps {
    node: SiteAdminPackageFields
}

const PackageNode: React.FunctionComponent<React.PropsWithChildren<PackageNodeProps>> = ({ node }) => {
    const packageRepository = node.repository

    return (
        <li className="list-group-item px-0 py-2">
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    {packageRepository ? (
                        <>
                            <ExternalRepositoryIcon externalRepo={packageRepository.externalRepository} />
                            <RepoLink repoName={node.name} to={packageRepository.url} />
                            <RepoMirrorInfo mirrorInfo={packageRepository.mirrorInfo} />
                        </>
                    ) : (
                        <>{node.name}</>
                    )}
                </div>
            </div>
            <div>
                {packageRepository?.mirrorInfo.lastError && (
                    <div className={styles.alertWrapper}>
                        <Alert variant="warning">
                            <Text className="font-weight-bold">Error syncing repository:</Text>
                            <Code className={styles.alertContent}>
                                {packageRepository.mirrorInfo.lastError.replaceAll('\r', '\n')}
                            </Code>
                        </Alert>
                    </div>
                )}
                {packageRepository?.mirrorInfo.isCorrupted && (
                    <div className={styles.alertWrapper}>
                        <Alert variant="danger">
                            Repository is corrupt. <Link to={`/${node.name}/-/settings/mirror`}>More details</Link>
                        </Alert>
                    </div>
                )}
            </div>
        </li>
    )
}

const EXTERNAL_SERVICE_KIND_TO_PACKAGE_REPO_REFERENCE_KIND: Partial<
    Record<
        ExternalServiceKind,
        {
            label: string
            value: PackageScheme
        }
    >
> = {
    [ExternalServiceKind.NPMPACKAGES]: {
        label: 'NPM',
        value: 'npm',
    },
    [ExternalServiceKind.GOMODULES]: {
        label: 'Go',
        value: 'go',
    },
    [ExternalServiceKind.JVMPACKAGES]: {
        label: 'JVM',
        value: 'semanticdb',
    },
    [ExternalServiceKind.RUBYPACKAGES]: {
        label: 'Ruby',
        value: 'scip-ruby',
    },
    [ExternalServiceKind.PYTHONPACKAGES]: {
        label: 'Python',
        value: 'python',
    },
    [ExternalServiceKind.RUSTPACKAGES]: {
        label: 'Rust',
        value: 'rust-analyzer',
    },
}

interface SiteAdminPackagesPageProps extends TelemetryProps {}

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
                label: 'Ecosystem',
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

    const {
        connection,
        error: packagesError,
        loading: packagesLoading,
        fetchMore,
        hasNextPage,
    } = useShowMorePagination<PackagesResult, PackagesVariables, SiteAdminPackageFields>({
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

    const error = extSvcError || packagesError
    const loading = extSvcLoading || packagesLoading

    return (
        <div>
            <PageTitle title="Packages - Admin" />
            <PageHeader
                path={[{ text: 'Packages' }]}
                headingElement="h2"
                description={
                    <>
                        Packages are synced from connected <Link to="/site-admin/external-services">package hosts</Link>
                        .
                    </>
                }
                className="mb-3"
            />

            <Container className="mb-3">
                {error && !loading && <ErrorAlert error={error} />}
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

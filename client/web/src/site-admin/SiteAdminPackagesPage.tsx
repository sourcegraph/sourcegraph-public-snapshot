import { useEffect, useMemo, useState } from 'react'

import { mdiBlockHelper, mdiCog, mdiDotsHorizontal } from '@mdi/js'
import { isEqual } from 'lodash'
import { useLocation, useNavigate } from 'react-router-dom'

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
    Icon,
    Menu,
    MenuButton,
    MenuList,
    MenuItem,
    MenuLink,
    Position,
} from '@sourcegraph/wildcard'

import { externalRepoIcon } from '../components/externalServices/externalServices'
import {
    buildFilterArgs,
    FilterControl,
    FilteredConnectionFilter,
    FilteredConnectionFilterValue,
} from '../components/FilteredConnection'
import { useShowMorePagination } from '../components/FilteredConnection/hooks/useShowMorePagination'
import { ConnectionSummary } from '../components/FilteredConnection/ui'
import { getFilterFromURL, getUrlQuery } from '../components/FilteredConnection/utils'
import { PageTitle } from '../components/PageTitle'
import {
    PackagesResult,
    PackagesVariables,
    SiteAdminPackageFields,
    ExternalServiceKindsVariables,
    ExternalServiceKindsResult,
    ExternalServiceKind,
    PackageRepoReferenceKind,
} from '../graphql-operations'

import { EXTERNAL_SERVICE_KINDS, PACKAGES_QUERY } from './backend'
import { RepoMirrorInfo } from './components/RepoMirrorInfo'
import { BlockPackagesModal } from './packages/BlockPackageModal'

import styles from './SiteAdminPackagesPage.module.scss'

const ExternalServicePackageMap: Partial<
    Record<
        ExternalServiceKind,
        {
            label: string
            value: PackageRepoReferenceKind
        }
    >
> = {
    [ExternalServiceKind.NPMPACKAGES]: {
        label: 'npm',
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

interface PackageNodeProps {
    node: SiteAdminPackageFields
    setSelectedPackage: (node: SiteAdminPackageFields) => void
}

const PackageNode: React.FunctionComponent<React.PropsWithChildren<PackageNodeProps>> = ({
    node,
    setSelectedPackage,
}) => {
    const PackageIconComponent = externalRepoIcon({ serviceType: node.kind })

    const packageRepository = node.repository

    return (
        <li className="list-group-item px-0 py-2">
            <div>
                <div className={styles.node}>
                    <div>
                        <Icon as={PackageIconComponent} aria-label="Package host logo" className="mr-2" />
                        {packageRepository ? (
                            <>
                                <RepoLink repoName={node.name} to={packageRepository.url} />
                                <RepoMirrorInfo mirrorInfo={packageRepository.mirrorInfo} />
                            </>
                        ) : (
                            <>{node.name}</>
                        )}
                    </div>
                    <div>
                        <Menu>
                            <MenuButton outline={true} aria-label="Package action">
                                <Icon svgPath={mdiDotsHorizontal} inline={false} aria-hidden={true} />
                            </MenuButton>
                            <MenuList position={Position.bottomEnd}>
                                {packageRepository?.mirrorInfo.cloned &&
                                    !packageRepository.mirrorInfo.lastError &&
                                    !packageRepository.mirrorInfo.cloneInProgress && (
                                        <MenuLink
                                            as={Link}
                                            to={`/${packageRepository.name}/-/settings`}
                                            className="p-2"
                                        >
                                            <Icon aria-hidden={true} svgPath={mdiCog} className="mr-1" />
                                            Settings
                                        </MenuLink>
                                    )}
                                <MenuItem
                                    as={Button}
                                    variant="danger"
                                    onSelect={() => setSelectedPackage(node)}
                                    className="p-2"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiBlockHelper} className="mr-1" />
                                    Block
                                </MenuItem>
                            </MenuList>
                        </Menu>
                    </div>
                </div>
                {packageRepository && (
                    <div>
                        {packageRepository.mirrorInfo.lastError && (
                            <div className={styles.alertWrapper}>
                                <Alert variant="warning">
                                    <Text className="font-weight-bold">Error syncing package:</Text>
                                    <Code className={styles.alertContent}>
                                        {packageRepository.mirrorInfo.lastError.replaceAll('\r', '\n')}
                                    </Code>
                                </Alert>
                            </div>
                        )}
                        {packageRepository.mirrorInfo.isCorrupted && (
                            <div className={styles.alertWrapper}>
                                <Alert variant="danger">
                                    Package is corrupt. <Link to={`/${node.name}/-/settings/mirror`}>More details</Link>
                                </Alert>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </li>
    )
}

interface SiteAdminPackagesPageProps extends TelemetryProps {}

const DEFAULT_FIRST = 15

/**
 * A page displaying the packages on this instance.
 */
export const SiteAdminPackagesPage: React.FunctionComponent<React.PropsWithChildren<SiteAdminPackagesPageProps>> = ({
    telemetryService,
}) => {
    const location = useLocation()
    const navigate = useNavigate()
    const [selectedPackage, setSelectedPackage] = useState<SiteAdminPackageFields | null>(null)

    useEffect(() => {
        telemetryService.logPageView('SiteAdminPackages')
    }, [telemetryService])

    const {
        loading: extSvcLoading,
        data: extSvcs,
        error: extSvcError,
    } = useQuery<ExternalServiceKindsResult, ExternalServiceKindsVariables>(EXTERNAL_SERVICE_KINDS, {})

    const ecosystemFilterValues = useMemo<FilteredConnectionFilterValue[]>(() => {
        const values = []

        for (const extSvc of extSvcs?.externalServices.nodes ?? []) {
            const packageRepoScheme = ExternalServicePackageMap[extSvc.kind]

            if (packageRepoScheme) {
                values.push({
                    ...packageRepoScheme,
                    args: { scheme: packageRepoScheme.value },
                })
            }
        }

        return values
    }, [extSvcs?.externalServices.nodes])

    const filters = useMemo<FilteredConnectionFilter[]>(
        () => [
            {
                id: 'ecosystem',
                label: 'Ecosystem',
                type: 'select',
                values: [
                    {
                        label: 'All',
                        value: 'all',
                        args: {},
                    },
                    ...ecosystemFilterValues,
                ],
            },
        ],
        [ecosystemFilterValues]
    )

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
            kind: null,
            after: null,
            first: DEFAULT_FIRST,
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
        <>
            {selectedPackage && (
                <BlockPackagesModal
                    node={selectedPackage}
                    filters={ecosystemFilterValues}
                    onDismiss={() => setSelectedPackage(null)}
                />
            )}
            <div>
                <PageTitle title="Packages - Admin" />
                <PageHeader
                    path={[{ text: 'Packages' }]}
                    headingElement="h2"
                    description={
                        <>
                            Packages are synced from connected{' '}
                            <Link to="/site-admin/external-services">code hosts</Link>.
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
                    <div className="d-flex align-items-end justify-content-between mt-3">
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
                        {connection && (
                            <ConnectionSummary
                                connection={connection}
                                connectionQuery={query}
                                hasNextPage={hasNextPage}
                                first={DEFAULT_FIRST}
                                noun="package"
                                pluralNoun="packages"
                                className="mb-0"
                            />
                        )}
                    </div>
                    {loading && !error && <LoadingSpinner className="d-block mx-auto mt-3" />}
                    {connection?.nodes && connection.nodes.length > 0 && (
                        <ul className="list-group list-group-flush mt-2">
                            {(connection?.nodes || []).map(node => (
                                <PackageNode node={node} key={node.id} setSelectedPackage={setSelectedPackage} />
                            ))}
                        </ul>
                    )}
                    {connection?.nodes && connection.totalCount !== connection.nodes.length && hasNextPage && (
                        <div>
                            <Button variant="link" size="sm" onClick={fetchMore}>
                                Show more
                            </Button>
                        </div>
                    )}
                </Container>
            </div>
        </>
    )
}

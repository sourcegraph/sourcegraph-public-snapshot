import { useCallback, useState } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { useQuery } from '@sourcegraph/http-client'
import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import {
    Button,
    Icon,
    Input,
    Label,
    Alert,
    Tooltip,
    useDebounce,
    LoadingSpinner,
    ErrorAlert,
    Select,
    Form,
} from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../../components/FilteredConnection'
import {
    SiteAdminPackageFields,
    PackageRepoReferenceKind,
    PackageRepoReferencesMatchingFilterResult,
    PackageRepoReferencesMatchingFilterVariables,
    PackageRepoMatchFields,
} from '../../../graphql-operations'
import { prettyBytesBigint } from '../../../util/prettyBytesBigint'
import { packageRepoFilterQuery } from '../backend'
import { BlockType } from '../modal-content/AddPackageFilterModalContent'

import { FilterPackagesActions } from './FilterPackagesActions'

import styles from '../modal-content/AddPackageFilterModalContent.module.scss'

export interface MultiPackageState {
    nameFilter: string
    ecosystem: PackageRepoReferenceKind
}

interface BaseMultiPackageFormProps {
    filters: FilteredConnectionFilterValue[]
    setType: (type: BlockType) => void
    onDismiss: () => void
    onSave: (state: MultiPackageState) => Promise<unknown>
}

interface AddMultiPackageFormProps extends BaseMultiPackageFormProps {
    node: SiteAdminPackageFields
}

interface EditMultiPackageFormProps extends BaseMultiPackageFormProps {
    initialState: MultiPackageState
}

type MultiPackageFormProps = AddMultiPackageFormProps | EditMultiPackageFormProps

export const MultiPackageForm: React.FunctionComponent<MultiPackageFormProps> = props => {
    const defaultNameFilter = 'initialState' in props ? props.initialState.nameFilter : '*'
    const defaultEcosystem = 'initialState' in props ? props.initialState.ecosystem : props.node.kind
    const [blockState, setBlockState] = useState<MultiPackageState>({
        nameFilter: defaultNameFilter,
        ecosystem: defaultEcosystem,
    })

    const [packageFetchLimit, setPackageFetchLimit] = useState(15)
    const query = useDebounce(blockState.nameFilter, 200)

    const { data, loading, error } = useQuery<
        PackageRepoReferencesMatchingFilterResult,
        PackageRepoReferencesMatchingFilterVariables
    >(packageRepoFilterQuery, {
        variables: {
            kind: blockState.ecosystem,
            filter: {
                nameFilter: {
                    packageGlob: query,
                },
            },
            first: packageFetchLimit,
        },
    })

    const packageCount = data?.packageRepoReferencesMatchingFilter.totalCount ?? 0

    // Limit fetching more than 1000 packages
    const nextFetchLimit = Math.min(packageCount, 1000)

    const isValid = useCallback((): boolean => {
        if (blockState.nameFilter === '') {
            return false
        }

        return true
    }, [blockState])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): Promise<unknown> => {
            event.preventDefault()

            if (!isValid()) {
                return Promise.resolve()
            }

            return props.onSave(blockState)
        },
        [blockState, isValid, props]
    )

    return (
        <>
            <Form onSubmit={handleSubmit} className="w-100 mb-3">
                <div>
                    <Label className="mb-2" id="package-name">
                        Name
                    </Label>
                    <div className={styles.inputRow}>
                        <Select
                            className={classNames('mr-1 mb-0', styles.ecosystemSelect)}
                            value={blockState.ecosystem}
                            onChange={event =>
                                setBlockState({
                                    ...blockState,
                                    ecosystem: event.target.value as PackageRepoReferenceKind,
                                })
                            }
                            required={true}
                            isCustomStyle={true}
                            aria-label="Ecosystem"
                        >
                            {props.filters.map(({ label, value }) => (
                                <option value={value} key={label}>
                                    {label}
                                </option>
                            ))}
                        </Select>
                        <Input
                            className="mr-2 flex-1"
                            value={blockState.nameFilter || ''}
                            required={true}
                            onChange={event => setBlockState({ ...blockState, nameFilter: event.target.value })}
                            placeholder="Example: @types/*"
                        />
                        <Tooltip content="Remove name filter">
                            <Button
                                className={classNames('text-danger', styles.inputRowButton)}
                                variant="icon"
                                onClick={() => props.setType('single')}
                            >
                                <Icon aria-hidden={true} svgPath={mdiClose} />
                            </Button>
                        </Tooltip>
                    </div>
                    <div className={styles.listContainer}>
                        <>
                            {loading ? (
                                <LoadingSpinner className="d-block mx-auto mt-3" />
                            ) : error ? (
                                <ErrorAlert error={error} className="mt-2" />
                            ) : data ? (
                                <div className="mt-2">
                                    <div className="d-flex justify-content-between text-muted">
                                        <span>
                                            {data.packageRepoReferencesMatchingFilter.totalCount === 1 ? (
                                                <>
                                                    {data.packageRepoReferencesMatchingFilter.totalCount} package
                                                    currently matches
                                                </>
                                            ) : (
                                                <>
                                                    {data.packageRepoReferencesMatchingFilter.totalCount} packages
                                                    currently match
                                                </>
                                            )}{' '}
                                            this filter
                                            {data.packageRepoReferencesMatchingFilter.nodes.length <
                                                data.packageRepoReferencesMatchingFilter.totalCount && (
                                                <>
                                                    {' '}
                                                    (showing only{' '}
                                                    {data.packageRepoReferencesMatchingFilter.nodes.length})
                                                </>
                                            )}
                                            :
                                        </span>
                                        {data.packageRepoReferencesMatchingFilter.nodes.length <
                                            data.packageRepoReferencesMatchingFilter.totalCount && (
                                            <Button
                                                variant="link"
                                                className="p-0 mr-3"
                                                onClick={() => setPackageFetchLimit(nextFetchLimit)}
                                            >
                                                <>
                                                    Show{' '}
                                                    {nextFetchLimit ===
                                                    data.packageRepoReferencesMatchingFilter.totalCount
                                                        ? 'all '
                                                        : { nextFetchLimit }}
                                                </>
                                            </Button>
                                        )}
                                    </div>
                                    <PackageList nodes={data.packageRepoReferencesMatchingFilter.nodes} />
                                </div>
                            ) : (
                                <></>
                            )}
                        </>
                    </div>
                </div>
                <div className="mt-3">
                    <Label className="mb-2">Version</Label>
                    <Alert variant="info">
                        All versions of all matching packages are blocked when using a name filter.
                    </Alert>
                </div>
                <FilterPackagesActions valid={isValid()} onDismiss={props.onDismiss} />
            </Form>
        </>
    )
}

interface PackageListProps {
    nodes: PackageRepoMatchFields[]
}
const PackageList: React.FunctionComponent<PackageListProps> = ({ nodes }) => {
    if (nodes.length === 0) {
        return (
            <Alert variant="warning" className="mt-1">
                This filter does not match any current package.
            </Alert>
        )
    }

    return (
        <ul className={classNames('list-group mt-1', styles.list)}>
            {nodes.map(node => (
                <li className="list-group-item" key={node.id}>
                    {node.repository ? (
                        <div className="d-flex justify-content-between">
                            <RepoLink repoName={node.name} to={node.repository.url} />
                            <small>Size: {prettyBytesBigint(BigInt(node.repository.mirrorInfo.byteSize))}</small>
                        </div>
                    ) : (
                        <>{node.name}</>
                    )}
                </li>
            ))}
        </ul>
    )
}

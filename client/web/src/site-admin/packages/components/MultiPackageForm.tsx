import { useCallback, useState } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

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
import { SiteAdminPackageFields, PackageRepoReferenceKind } from '../../../graphql-operations'
import { prettyBytesBigint } from '../../../util/prettyBytesBigint'
import { useMatchingPackages } from '../hooks/useMatchingPackages'
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
    const query = useDebounce(blockState.nameFilter, 200)

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
                        <PackageList query={query} blockState={blockState} />
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
    blockState: MultiPackageState
    query: string
}
const PackageList: React.FunctionComponent<PackageListProps> = ({ blockState, query }) => {
    const [packageFetchLimit, setPackageFetchLimit] = useState(15)
    const { nodes, totalCount, loading, error } = useMatchingPackages({
        kind: blockState.ecosystem,
        filter: {
            nameFilter: {
                packageGlob: query,
            },
        },
        first: packageFetchLimit,
    })

    // Limit fetching more than 1000 packages
    const nextFetchLimit = Math.min(totalCount, 1000)

    if (loading) {
        return <LoadingSpinner className="d-block mx-auto mt-2" />
    }

    if (error) {
        return <ErrorAlert error={error} className="mt-2" />
    }

    if (nodes.length === 0) {
        return (
            <Alert variant="warning" className="mt-2">
                This filter does not match any current package.
            </Alert>
        )
    }

    return (
        <div className="mt-2">
            <div className="d-flex justify-content-between text-muted">
                <span>
                    {totalCount === 1 ? (
                        <>{totalCount} package currently matches</>
                    ) : (
                        <>{totalCount} packages currently match</>
                    )}{' '}
                    this filter
                    {nodes.length < totalCount && <> (showing only {nodes.length})</>}:
                </span>
                {nodes.length < totalCount && (
                    <Button variant="link" className="p-0 mr-3" onClick={() => setPackageFetchLimit(nextFetchLimit)}>
                        <>Show {nextFetchLimit === totalCount ? 'all ' : { nextFetchLimit }}</>
                    </Button>
                )}
            </div>
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
        </div>
    )
}

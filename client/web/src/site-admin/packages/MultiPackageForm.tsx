import { useCallback, useState } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { useMutation, useQuery } from '@sourcegraph/http-client'
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

import { FilteredConnectionFilterValue } from '../../components/FilteredConnection'
import {
    SiteAdminPackageFields,
    PackageRepoReferenceKind,
    PackageRepoReferencesMatchingFilterResult,
    PackageRepoReferencesMatchingFilterVariables,
    PackageMatchBehaviour,
    AddPackageRepoFilterResult,
    AddPackageRepoFilterVariables,
    PackageRepoMatchFields,
} from '../../graphql-operations'
import { prettyBytesBigint } from '../../util/prettyBytesBigint'

import { addPackageRepoFilterMutation, packageRepoFilterQuery } from './backend'
import { BlockPackageActions } from './BlockPackageActions'
import { BlockType } from './BlockPackageModal'

import styles from './BlockPackageModal.module.scss'

interface MultiPackageState {
    nameFilter: string
    ecosystem: PackageRepoReferenceKind
}

interface MultiPackageFormProps {
    node: SiteAdminPackageFields
    filters: FilteredConnectionFilterValue[]
    setType: (type: BlockType) => void
    onDismiss: () => void
}

export const MultiPackageForm: React.FunctionComponent<MultiPackageFormProps> = ({
    node,
    filters,
    setType,
    onDismiss,
}) => {
    const [blockState, setBlockState] = useState<MultiPackageState>({
        nameFilter: '*',
        ecosystem: node.kind,
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

    const [submitPackageFilter, submitPackageFilterResponse] = useMutation<
        AddPackageRepoFilterResult,
        AddPackageRepoFilterVariables
    >(addPackageRepoFilterMutation, {})

    const packageCount = data?.packageRepoReferencesMatchingFilter.totalCount ?? 0

    // Limit fetching more than 1000 packages
    const nextFetchLimit = Math.min(packageCount, 1000)

    const isValid = useCallback((): boolean => {
        if (packageCount === 0) {
            return false
        }

        return true
    }, [packageCount])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()

            if (!isValid()) {
                return
            }

            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            submitPackageFilter({
                variables: {
                    kind: blockState.ecosystem,
                    behaviour: PackageMatchBehaviour.BLOCK,
                    filter: {
                        nameFilter: {
                            packageGlob: blockState.nameFilter,
                        },
                    },
                },
            })
        },
        [blockState, isValid, submitPackageFilter]
    )

    const getSubmitText = useCallback((): string => {
        if (packageCount === 1) {
            const packageNode = data?.packageRepoReferencesMatchingFilter.nodes[0]
            return `Block ${packageNode?.name}`
        }

        if (packageCount > 1) {
            return `Block ${packageCount} packages`
        }

        return 'Block packages'
    }, [data?.packageRepoReferencesMatchingFilter.nodes, packageCount])

    return (
        <>
            <Form onSubmit={handleSubmit} className="w-100 mb-3">
                <div>
                    <Label className="mb-2" id="package-name">
                        Name
                    </Label>
                    <div className={styles.inputRow}>
                        <Select
                            className={classNames('mr-1 mb-0', styles.select)}
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
                            {filters.map(({ label, value }) => (
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
                                onClick={() => setType('single')}
                            >
                                <Icon aria-hidden={true} svgPath={mdiClose} />
                            </Button>
                        </Tooltip>
                    </div>
                    <div className={styles.listContainer}>
                        <>
                            {loading || !data ? (
                                <LoadingSpinner className="d-block mx-auto mt-3" />
                            ) : error ? (
                                <ErrorAlert error={error} className="mt-2" />
                            ) : (
                                <div className="mt-2">
                                    <div className="d-flex justify-content-between text-muted">
                                        <span>
                                            {data.packageRepoReferencesMatchingFilter.totalCount === 1 ? (
                                                <>
                                                    {data.packageRepoReferencesMatchingFilter.totalCount} package
                                                    matches
                                                </>
                                            ) : (
                                                <>
                                                    {data.packageRepoReferencesMatchingFilter.totalCount} packages match
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
                <BlockPackageActions
                    submitText={getSubmitText()}
                    valid={isValid()}
                    error={submitPackageFilterResponse.error}
                    loading={submitPackageFilterResponse.loading}
                    onDismiss={onDismiss}
                />
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
                This filter does not match any package.
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

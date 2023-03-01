import { useState } from 'react'

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
} from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../components/FilteredConnection'
import { SiteAdminPackageFields, PackageRepoReferenceKind } from '../../graphql-operations'
import { prettyBytesBigint } from '../../util/prettyBytesBigint'

import { BlockType } from './BlockPackageModal'
import { usePackageRepoMatchesMock } from './mocks'

import styles from './BlockPackageModal.module.scss'

interface MultiPackageState {
    namePattern: string | null
    ecosystem: PackageRepoReferenceKind
}

interface MultiPackageFormProps {
    node: SiteAdminPackageFields
    filters: FilteredConnectionFilterValue[]
    setType: (type: BlockType) => void
}

export const MultiPackageForm: React.FunctionComponent<MultiPackageFormProps> = ({ node, filters, setType }) => {
    const [blockState, setBlockState] = useState<MultiPackageState>({
        namePattern: null,
        ecosystem: node.scheme,
    })
    const query = useDebounce(blockState.namePattern, 200)

    return (
        <>
            <div className="w-100 mb-3">
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
                            value={blockState.namePattern || ''}
                            onChange={event => setBlockState({ ...blockState, namePattern: event.target.value })}
                            placeholder="Example: @types/*"
                        />
                        <Tooltip content="Remove block pattern">
                            <Button
                                className={classNames('text-danger', styles.inputRowButton)}
                                variant="icon"
                                onClick={() => setType('single')}
                            >
                                <Icon aria-hidden={true} svgPath={mdiClose} />
                            </Button>
                        </Tooltip>
                    </div>
                </div>
                {query !== null && <PackageList namePattern={query} ecosystem={blockState.ecosystem} />}
            </div>
        </>
    )
}

interface PackageListProps {
    namePattern: string
    ecosystem: PackageRepoReferenceKind
}
const PackageList: React.FunctionComponent<PackageListProps> = ({ namePattern, ecosystem }) => {
    const [packageFetchLimit, setPackageFetchLimit] = useState(15)
    const { data, loading, error } = usePackageRepoMatchesMock({
        variables: {
            scheme: ecosystem,
            filter: {
                nameMatcher: {
                    packageGlob: namePattern,
                },
            },
            first: packageFetchLimit,
        },
    })

    if (loading || !data) {
        return <LoadingSpinner className="d-block mx-auto mt-3" />
    }

    if (error) {
        return <ErrorAlert error={error} />
    }

    const { totalCount, nodes } = data.packageReposMatches

    // Limit fetching more than 1000 packages
    const nextFetchLimit = Math.min(totalCount, 1000)

    if (nodes.length === 0) {
        return (
            <Alert variant="warning" className="mt-3">
                This pattern does not match any package.
            </Alert>
        )
    }

    return (
        <div className="mt-2">
            <div className="d-flex justify-content-between">
                <span>
                    {totalCount === 1 ? <>{totalCount} package matches</> : <>{totalCount} packages match</>} this
                    pattern
                    {nodes.length < totalCount && <> (showing only {nodes.length})</>}:
                </span>
                {nodes.length < totalCount && (
                    <Button variant="link" className="p-0 mr-3" onClick={() => setPackageFetchLimit(nextFetchLimit)}>
                        Show {nextFetchLimit === totalCount ? 'all ' : { nextFetchLimit }}
                    </Button>
                )}
            </div>
            <ul className={classNames('list-group', styles.list)}>
                {data.packageReposMatches.nodes.map(node => (
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

import { useState, useCallback } from 'react'

import { mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { toRepoURL } from '@sourcegraph/shared/src/util/url'
import {
    Button,
    Icon,
    Input,
    Label,
    Tooltip,
    Select,
    LoadingSpinner,
    ErrorAlert,
    Badge,
    useDebounce,
    Form,
    Alert,
    Link,
} from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../../components/FilteredConnection'
import { SiteAdminPackageFields, PackageRepoReferenceKind, PackageRepoMatchFields } from '../../../graphql-operations'
import { useMatchingPackages } from '../hooks/useMatchingPackages'
import { useMatchingVersions } from '../hooks/useMatchingVersions'
import { BlockType } from '../modal-content/AddPackageFilterModalContent'

import { FilterPackagesActions } from './FilterPackagesActions'

import styles from '../modal-content/AddPackageFilterModalContent.module.scss'

export interface SinglePackageState {
    name: string
    ecosystem: PackageRepoReferenceKind
    versionFilter: string
}

interface BaseSinglePackageFormProps {
    filters: FilteredConnectionFilterValue[]
    setType: (type: BlockType) => void
    onDismiss: () => void
    onSave: (state: SinglePackageState) => Promise<unknown>
}

interface AddSinglePackageFormProps extends BaseSinglePackageFormProps {
    node: SiteAdminPackageFields
}

interface EditSinglePackageFormProps extends BaseSinglePackageFormProps {
    initialState: SinglePackageState
}

type SinglePackageFormProps = AddSinglePackageFormProps | EditSinglePackageFormProps

export const SinglePackageForm: React.FunctionComponent<SinglePackageFormProps> = props => {
    const defaultName = 'initialState' in props ? props.initialState.name : props.node.name
    const defaultVersionFilter = 'initialState' in props ? props.initialState.versionFilter : '*'
    const defaultEcosystem = 'initialState' in props ? props.initialState.ecosystem : props.node.kind

    const [blockState, setBlockState] = useState<SinglePackageState>({
        name: defaultName,
        versionFilter: defaultVersionFilter,
        ecosystem: defaultEcosystem,
    })
    const versionQuery = useDebounce(blockState.versionFilter, 200)

    const { nodes } = useMatchingPackages({
        first: 1,
        kind: blockState.ecosystem,
        filter: {
            nameFilter: {
                packageGlob: defaultName,
            },
        },
    })

    const isValid = useCallback((): boolean => {
        if (blockState.versionFilter === '') {
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
        <Form onSubmit={handleSubmit} className="w-100 mb-3">
            <div>
                <Label className="mb-2" id="package-name">
                    Name
                </Label>
                <div className={styles.inputRow}>
                    <Select
                        name="single-ecosystem-select"
                        className={classNames('mr-1 mb-0', styles.ecosystemSelect)}
                        value={blockState.ecosystem}
                        disabled={true}
                        isCustomStyle={true}
                        required={true}
                        aria-label="Ecosystem"
                    >
                        {props.filters.map(({ label, value }) => (
                            <option value={value} key={label}>
                                {label}
                            </option>
                        ))}
                    </Select>
                    <Input
                        name="single-package-input"
                        className="mr-2 flex-1"
                        value={blockState.name}
                        disabled={true}
                        required={true}
                        aria-labelledby="package-name"
                    />
                    <Tooltip content="Block multiple packages at once">
                        <Button
                            className={styles.inputRowButton}
                            variant="secondary"
                            outline={true}
                            onClick={() => props.setType('multiple')}
                        >
                            <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-1" />
                            Filter
                        </Button>
                    </Tooltip>
                </div>
            </div>
            <div className="mt-3">
                <Label className="mb-2" id="package-version">
                    Version
                </Label>
                <div className={styles.inputRow}>
                    <Input
                        name="multi-version-input"
                        aria-labelledby="package-version"
                        className="mr-2 flex-1"
                        value={blockState.versionFilter || ''}
                        placeholder="e.g. v1.*"
                        required={true}
                        onChange={event => setBlockState({ ...blockState, versionFilter: event.target.value })}
                    />
                </div>
            </div>
            <VersionList blockState={blockState} query={versionQuery} node={nodes[0]} />
            <FilterPackagesActions valid={isValid()} onDismiss={props.onDismiss} />
        </Form>
    )
}

interface VersionListProps {
    node?: PackageRepoMatchFields
    blockState: SinglePackageState
    query: string
}
const VersionList: React.FunctionComponent<VersionListProps> = ({ blockState, query, node }) => {
    const [versionFetchLimit, setVersionFetchLimit] = useState(15)
    const { versions, totalCount, loading, error } = useMatchingVersions({
        kind: blockState.ecosystem,
        filter: {
            versionFilter: {
                packageName: blockState.name,
                versionGlob: query,
            },
        },
        first: versionFetchLimit,
    })

    // Limit fetching more than 1000 versions
    const nextFetchLimit = Math.min(totalCount, 1000)

    if (loading) {
        return <LoadingSpinner className="d-block mx-auto mt-2" />
    }

    if (error) {
        return <ErrorAlert error={error} className="mt-2" />
    }

    if (versions.length === 0) {
        return (
            <Alert variant="warning" className="mt-2">
                This filter does not match any current version.
            </Alert>
        )
    }

    return (
        <div className="mt-2">
            <div className="d-flex justify-content-between text-muted">
                <span>
                    {totalCount === 1 ? (
                        <>{totalCount} version currently matches</>
                    ) : (
                        <>{totalCount} versions currently match</>
                    )}{' '}
                    this filter
                    {versions.length < totalCount && <> (showing only {versions.length})</>}:
                </span>
                {versions.length < totalCount && (
                    <Button variant="link" className="p-0 mr-3" onClick={() => setVersionFetchLimit(nextFetchLimit)}>
                        <>Show {nextFetchLimit === totalCount ? 'all ' : { nextFetchLimit }}</>
                    </Button>
                )}
            </div>
            <ul className={classNames('list-group mt-1', styles.list)}>
                {versions.map(version => (
                    <li className="list-group-item" key={version}>
                        {node?.repository ? (
                            <div className="d-flex justify-content-between">
                                <Link
                                    to={toRepoURL({
                                        repoName: node.repository.name,
                                        revision: `v${version}`,
                                    })}
                                >
                                    <Badge className="px-2 py-0" as="code">
                                        {version}
                                    </Badge>
                                </Link>
                            </div>
                        ) : (
                            <Badge className="px-2 py-0" as="code">
                                {version}
                            </Badge>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    )
}

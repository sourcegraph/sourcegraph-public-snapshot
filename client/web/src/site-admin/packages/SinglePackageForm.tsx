import { useState } from 'react'

import { mdiClose, mdiPlus } from '@mdi/js'
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
    Link,
    Badge,
    useDebounce,
} from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../components/FilteredConnection'
import { PackageRepoReferenceKind, SiteAdminPackageFields } from '../../graphql-operations'

import { BlockType } from './BlockPackageModal'
import { usePackageRepoMatchesMock } from './mocks'

import styles from './BlockPackageModal.module.scss'

interface SinglePackageSingleVersionState {
    name: string
    version: string
}

interface SinglePackageMultiVersionState {
    name: string
    versionPattern: string | null
}

type SinglePackageState = SinglePackageSingleVersionState | SinglePackageMultiVersionState

interface SinglePackageFormProps {
    node: SiteAdminPackageFields
    filters: FilteredConnectionFilterValue[]
    setType: (type: BlockType) => void
}

export const SinglePackageForm: React.FunctionComponent<SinglePackageFormProps> = ({ node, filters, setType }) => {
    // TODO: Is this sorted?
    const defaultVersion = node.versions[0].version

    const [blockState, setBlockState] = useState<SinglePackageState>({
        name: node.name,
        version: defaultVersion,
    })

    const versionQuery = useDebounce('versionPattern' in blockState ? blockState.versionPattern : null, 200)

    return (
        <div className="w-100 mb-3">
            <div>
                <Label className="mb-2" id="package-name">
                    Name
                </Label>
                <div className={styles.inputRow}>
                    <Select
                        className={classNames('mr-1 mb-0', styles.select)}
                        value={node.scheme}
                        disabled={true}
                        isCustomStyle={true}
                        aria-label="Ecosystem"
                    >
                        {filters.map(({ label, value }) => (
                            <option value={value} key={label}>
                                {label}
                            </option>
                        ))}
                    </Select>
                    <Input className="mr-2 flex-1" value={node.name} disabled={true} aria-labelledby="package-name" />
                    <Tooltip content="Block multiple packages at once">
                        <Button
                            className={styles.inputRowButton}
                            variant="secondary"
                            outline={true}
                            size="sm"
                            onClick={() => setType('multiple')}
                        >
                            <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-1" />
                            Pattern
                        </Button>
                    </Tooltip>
                </div>
            </div>
            <div className="mt-3">
                <Label className="mb-2" id="package-version">
                    Version
                </Label>
                <div className={styles.inputRow}>
                    {'version' in blockState && (
                        <>
                            <Select
                                aria-labelledby="package-version"
                                value={blockState.version}
                                onChange={event =>
                                    setBlockState({
                                        ...blockState,
                                        version: event.target.value,
                                    })
                                }
                                className="mr-2 mb-0 flex-1"
                                isCustomStyle={true}
                            >
                                {node.versions.map(({ id, version }) => (
                                    <option value={version} key={id}>
                                        {version}
                                    </option>
                                ))}
                            </Select>
                            <Tooltip content="Block multiple versions at once">
                                <Button
                                    className={styles.inputRowButton}
                                    variant="secondary"
                                    outline={true}
                                    size="sm"
                                    onClick={() => setBlockState({ name: node.name, versionPattern: null })}
                                >
                                    <Icon aria-hidden={true} svgPath={mdiPlus} className="mr-1" />
                                    Pattern
                                </Button>
                            </Tooltip>
                        </>
                    )}
                    {'versionPattern' in blockState && (
                        <>
                            <Input
                                aria-labelledby="package-version"
                                className="mr-2 flex-1"
                                value={blockState.versionPattern || ''}
                                placeholder="e.g. v1.*"
                                onChange={event => setBlockState({ ...blockState, versionPattern: event.target.value })}
                            />
                            <Tooltip content="Remove version pattern">
                                <Button
                                    variant="icon"
                                    className="text-danger"
                                    onClick={() =>
                                        setBlockState({
                                            name: node.name,
                                            version: defaultVersion,
                                        })
                                    }
                                >
                                    <Icon aria-hidden={true} svgPath={mdiClose} />
                                </Button>
                            </Tooltip>
                        </>
                    )}
                </div>
            </div>
            {'versionPattern' in blockState && versionQuery !== null && (
                <VersionList name={blockState.name} versionPattern={versionQuery} />
            )}
        </div>
    )
}

interface VersionListProps {
    name: string
    versionPattern: string
}
const VersionList: React.FunctionComponent<VersionListProps> = ({ name, versionPattern }) => {
    const { data, loading, error } = usePackageRepoMatchesMock({
        variables: {
            scheme: PackageRepoReferenceKind.NPMPACKAGES,
            filter: {
                versionMatcher: {
                    packageName: name,
                    versionGlob: versionPattern,
                },
            },
            first: 100, // TODO: Limit?
        },
    })

    if (loading || !data) {
        return <LoadingSpinner className="d-block mx-auto mt-3" />
    }

    if (error) {
        return <ErrorAlert error={error} />
    }

    const packageNode = data.packageReposMatches.nodes[0]
    const versionCount = packageNode.versions.length

    return (
        <div className="mt-2">
            <div className="d-flex justify-content-between">
                <span>
                    {versionCount === 1 ? <>{versionCount} version matches</> : <>{versionCount} versions match</>} this
                    pattern
                </span>
            </div>
            <ul className={classNames('list-group', styles.list)}>
                {packageNode.versions.map(({ id, version }) => (
                    <li className="list-group-item" key={id}>
                        {packageNode.repository ? (
                            <div className="d-flex justify-content-between">
                                <Link
                                    to={toRepoURL({
                                        repoName: packageNode.repository.name,
                                        revision: version,
                                    })}
                                >
                                    <Badge className="px-2 py-0" as="code">
                                        {version}
                                    </Badge>
                                </Link>
                            </div>
                        ) : (
                            <>{packageNode.name}</>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    )
}

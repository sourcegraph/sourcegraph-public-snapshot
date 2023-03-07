import { useState, useCallback } from 'react'

import { mdiClose, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { useMutation, useQuery } from '@sourcegraph/http-client'
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
    Form,
} from '@sourcegraph/wildcard'

import { FilteredConnectionFilterValue } from '../../components/FilteredConnection'
import {
    PackageRepoReferencesMatchingFilterResult,
    PackageRepoReferencesMatchingFilterVariables,
    SiteAdminPackageFields,
    PackageRepoMatchFields,
    AddPackageRepoFilterResult,
    AddPackageRepoFilterVariables,
    PackageMatchBehaviour,
} from '../../graphql-operations'

import { addPackageRepoFilterMutation, packageRepoFilterQuery } from './backend'
import { BlockType } from './BlockPackageModal'
import { BlockPackageActions } from './BlockPackageActions'

import styles from './BlockPackageModal.module.scss'

interface SinglePackageSingleVersionState {
    name: string
    version: string
}

interface SinglePackageMultiVersionState {
    name: string
    versionPattern: string
}

type SinglePackageState = SinglePackageSingleVersionState | SinglePackageMultiVersionState

interface SinglePackageFormProps {
    node: SiteAdminPackageFields
    filters: FilteredConnectionFilterValue[]
    setType: (type: BlockType) => void
    onDismiss: () => void
}

export const SinglePackageForm: React.FunctionComponent<SinglePackageFormProps> = ({
    node,
    filters,
    setType,
    onDismiss,
}) => {
    // TODO: Is this sorted?
    const defaultVersion = node.versions[0].version

    const [blockState, setBlockState] = useState<SinglePackageState>({
        name: node.name,
        version: defaultVersion,
    })
    const versionQuery = useDebounce('versionPattern' in blockState ? blockState.versionPattern : null, 200)

    const versionPatternResponse = useQuery<
        PackageRepoReferencesMatchingFilterResult,
        PackageRepoReferencesMatchingFilterVariables
    >(packageRepoFilterQuery, {
        variables: {
            kind: node.kind,
            filter: {
                versionFilter: {
                    packageName: blockState.name,
                    versionGlob: (blockState as SinglePackageMultiVersionState).versionPattern,
                },
            },
            first: 100, // TODO: Limit?
        },
        skip: !('versionPattern' in blockState),
    })

    const [submitPackageFilter, submitPackageFilterResponse] = useMutation<
        AddPackageRepoFilterResult,
        AddPackageRepoFilterVariables
    >(addPackageRepoFilterMutation, {})

    const versionNode = versionPatternResponse.data?.packageRepoReferencesMatchingFilter?.nodes?.[0]
    const versionCount = versionNode?.versions?.length || 0

    const isValid = useCallback((): boolean => {
        if ('version' in blockState && blockState.version === '') {
            return false
        }

        if ('versionPattern' in blockState && versionCount === 0) {
            return false
        }

        return true
    }, [blockState, versionCount])

    const handleSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()

            if (!isValid()) {
                return
            }

            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            submitPackageFilter({
                variables: {
                    kind: node.kind,
                    behaviour: PackageMatchBehaviour.BLOCK,
                    filter: {
                        versionFilter: {
                            packageName: blockState.name,
                            versionGlob:
                                'versionPattern' in blockState ? blockState.versionPattern : blockState.version,
                        },
                    },
                },
            })
        },
        [blockState, isValid, node.kind, submitPackageFilter]
    )

    const getSubmitText = useCallback((): string => {
        if ('version' in blockState) {
            return `Block ${blockState.name}@${blockState.version}`
        }

        if (versionNode && versionCount === 1) {
            return `Block ${blockState.name}@${versionNode.versions[0].version}`
        }

        if (versionNode && versionCount > 1) {
            return `Block ${blockState.name} (${versionCount} versions)`
        }

        return `Block ${blockState.name}`
    }, [blockState, versionCount, versionNode])

    return (
        <Form onSubmit={handleSubmit} className="w-100 mb-3">
            <div>
                <Label className="mb-2" id="package-name">
                    Name
                </Label>
                <div className={styles.inputRow}>
                    <Select
                        name="single-ecosystem-select"
                        className={classNames('mr-1 mb-0', styles.select)}
                        value={node.kind}
                        disabled={true}
                        isCustomStyle={true}
                        required={true}
                        aria-label="Ecosystem"
                    >
                        {filters.map(({ label, value }) => (
                            <option value={value} key={label}>
                                {label}
                            </option>
                        ))}
                    </Select>
                    <Input
                        name="single-package-input"
                        className="mr-2 flex-1"
                        value={node.name}
                        disabled={true}
                        required={true}
                        aria-labelledby="package-name"
                    />
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
                                name="single-version-select"
                                aria-labelledby="package-version"
                                value={blockState.version}
                                onChange={event =>
                                    setBlockState({
                                        ...blockState,
                                        version: event.target.value,
                                    })
                                }
                                className="mr-2 mb-0 flex-1"
                                required={true}
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
                                    onClick={() => setBlockState({ name: node.name, versionPattern: '*' })}
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
                                name="multi-version-input"
                                aria-labelledby="package-version"
                                className="mr-2 flex-1"
                                value={blockState.versionPattern || ''}
                                placeholder="e.g. v1.*"
                                required={true}
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
                <>
                    {versionPatternResponse.loading || !versionPatternResponse.data ? (
                        <LoadingSpinner className="d-block mx-auto mt-3" />
                    ) : versionPatternResponse.error ? (
                        <ErrorAlert error={versionPatternResponse.error} />
                    ) : (
                        <VersionList node={versionPatternResponse.data.packageRepoReferencesMatchingFilter.nodes[0]} />
                    )}
                </>
            )}
            <BlockPackageActions
                submitText={getSubmitText()}
                valid={isValid()}
                error={submitPackageFilterResponse.error}
                loading={submitPackageFilterResponse.loading}
                onDismiss={onDismiss}
            />
        </Form>
    )
}

interface VersionListProps {
    node: PackageRepoMatchFields
}
const VersionList: React.FunctionComponent<VersionListProps> = ({ node }) => {
    const versionCount = node.versions.length

    return (
        <div className="mt-2">
            <div className="text-muted">
                {versionCount === 1 ? <>{versionCount} version matches</> : <>{versionCount} versions match</>} this
                pattern
            </div>
            <ul className={classNames('list-group mt-1', styles.list)}>
                {node.versions.map(({ id, version }) => (
                    <li className="list-group-item" key={id}>
                        {node.repository ? (
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

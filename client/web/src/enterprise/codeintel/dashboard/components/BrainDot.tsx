import React, { useMemo } from 'react'

import { mdiArrowRightThin, mdiBrain } from '@mdi/js'
import classNames from 'classnames'

import {
    Button,
    Position,
    Icon,
    Link,
    LoadingSpinner,
    Menu,
    MenuButton,
    MenuDivider,
    MenuHeader,
    MenuList,
    Tooltip,
    RadioButton,
    useSessionStorage,
    Code,
} from '@sourcegraph/wildcard'

import { INDEX_COMPLETED_STATES, INDEX_FAILURE_STATES } from '../constants'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'
import { useVisibleIndexes } from '../hooks/useVisibleIndexes'

import { getIndexerKey, getIndexRoot, sanitizePath } from './tree/util'

import styles from './BrainDot.module.scss'

export interface BrainDotProps {
    repoName: string
    commit: string
    path?: string
}

export const BrainDot: React.FunctionComponent<BrainDotProps> = ({ repoName, commit, path }) => {
    const { data: statusData } = useRepoCodeIntelStatus({ repository: repoName })

    const indexes = useMemo(() => {
        if (!statusData) {
            return []
        }

        return statusData.summary.recentActivity
    }, [statusData])

    const suggestedIndexers = useMemo(() => {
        if (!statusData) {
            return []
        }

        return statusData.summary.availableIndexers
            .flatMap(({ rootsWithKeys, indexer }) =>
                rootsWithKeys.map(({ root, comparisonKey }) => ({ root, comparisonKey, ...indexer }))
            )
            .filter(
                ({ root, key }) =>
                    !indexes.some(index => getIndexRoot(index) === sanitizePath(root) && getIndexerKey(index) === key)
            )
    }, [statusData, indexes])

    const dotStyle = useMemo((): string => {
        if (!indexes || !suggestedIndexers) {
            return ''
        }

        const numCompletedIndexes = indexes.filter(index => INDEX_COMPLETED_STATES.has(index.state)).length
        const numFailedIndexes = indexes.filter(index => INDEX_FAILURE_STATES.has(index.state)).length
        const numUnconfiguredProjects = suggestedIndexers.length

        return numFailedIndexes > 0
            ? styles.braindotDanger
            : numUnconfiguredProjects > 0
            ? styles.braindotWarning
            : numCompletedIndexes > 0
            ? styles.braindotSuccess
            : ''
    }, [indexes, suggestedIndexers])

    const { data: visibleIndexes, loading: visibleIndexesLoading } = useVisibleIndexes({
        repository: repoName,
        commit,
        path: path ?? '',
    })

    const [indexIDsForSnapshotData, setIndexIDForSnapshotData] = useSessionStorage<{
        [repoName: string]: string | undefined
    }>('blob.preciseIndexIDForSnapshotData', {})
    let visibleIndexID = indexIDsForSnapshotData[repoName]

    if (!visibleIndexes?.some(value => value.id === visibleIndexID)) {
        visibleIndexID = undefined
    }

    // TODO(nsc) - add a feature flag to enable nerd controls
    // https://github.com/sourcegraph/sourcegraph/pull/49128/files#diff-04df7090c83826679f92f4ee2881b626422057a8e6b59750937e2888d74e411cL152
    const forNerds = true

    return forNerds ? (
        <Menu>
            <>
                <Tooltip content="View code intelligence summary">
                    <MenuButton
                        className={classNames('text-decoration-none', styles.braindot, dotStyle)}
                        aria-label="Code graph"
                    >
                        <Icon aria-hidden={true} svgPath={mdiBrain} />
                    </MenuButton>
                </Tooltip>
                <MenuList position={Position.bottomEnd} className={styles.dropdownMenu}>
                    <MenuHeader>
                        Click to view code intelligence summary
                        <span className="float-right">
                            <Tooltip content="View code intelligence summary">
                                <Link to={`/${repoName}/-/code-graph/dashboard`}>
                                    <Icon aria-hidden={true} svgPath={mdiArrowRightThin} />
                                </Link>
                            </Tooltip>
                        </span>
                    </MenuHeader>

                    <MenuDivider />

                    {visibleIndexesLoading && <LoadingSpinner className="mx-2" />}
                    {visibleIndexes && visibleIndexes.length > 0 && (
                        <MenuHeader>
                            <Tooltip content="Not intended for regular use">
                                <span>Display debug information for uploaded index.</span>
                            </Tooltip>
                            {[
                                <RadioButton
                                    id="none"
                                    key="none"
                                    name="none"
                                    label="None"
                                    wrapperClassName="py-1"
                                    checked={visibleIndexID === undefined}
                                    onChange={() => {
                                        delete indexIDsForSnapshotData[repoName]
                                        setIndexIDForSnapshotData(indexIDsForSnapshotData)
                                    }}
                                />,
                                ...visibleIndexes.map(index => (
                                    <Tooltip content={`Uploaded at ${index.uploadedAt}`} key={index.id}>
                                        <RadioButton
                                            id={index.id}
                                            name={index.id}
                                            checked={visibleIndexID === index.id}
                                            wrapperClassName="py-1"
                                            label={
                                                <>
                                                    Index at <Code>{index.inputCommit.slice(0, 7)}</Code>
                                                </>
                                            }
                                            onChange={() => {
                                                indexIDsForSnapshotData[repoName] = index.id
                                                setIndexIDForSnapshotData(indexIDsForSnapshotData)
                                            }}
                                        />
                                    </Tooltip>
                                )),
                            ]}
                        </MenuHeader>
                    )}
                    {(visibleIndexes?.length ?? 0) === 0 && !visibleIndexesLoading && (
                        <MenuHeader>No precise indexes to display debug information for.</MenuHeader>
                    )}
                </MenuList>
            </>
        </Menu>
    ) : (
        <Tooltip content="View code intelligence summary">
            <Link to={`/${repoName}/-/code-graph/dashboard`}>
                <Button
                    className={classNames('text-decoration-none', styles.braindot, dotStyle)}
                    aria-label="Code graph"
                >
                    <Icon aria-hidden={true} svgPath={mdiBrain} />
                </Button>
            </Link>
        </Tooltip>
    )
}

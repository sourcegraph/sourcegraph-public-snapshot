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
} from '@sourcegraph/wildcard'

import { INDEX_COMPLETED_STATES, INDEX_FAILURE_STATES } from '../constants'
import { useRepoCodeIntelStatus } from '../hooks/useRepoCodeIntelStatus'

import { getIndexerKey, getIndexRoot, sanitizePath } from './tree/util'

import styles from './BrainDot.module.scss'

export interface BrainDotProps {
    repoName: string
}

export const BrainDot: React.FunctionComponent<BrainDotProps> = ({ repoName }) => {
    const { data } = useRepoCodeIntelStatus({ repository: repoName })

    const indexes = useMemo(() => {
        if (!data) {
            return []
        }

        return data.summary.recentActivity
    }, [data])

    const suggestedIndexers = useMemo(() => {
        if (!data) {
            return []
        }

        return data.summary.availableIndexers
            .flatMap(({ rootsWithKeys, indexer }) =>
                rootsWithKeys.map(({ root, comparisonKey }) => ({ root, comparisonKey, ...indexer }))
            )
            .filter(
                ({ root, key }) =>
                    !indexes.some(index => getIndexRoot(index) === sanitizePath(root) && getIndexerKey(index) === key)
            )
    }, [data, indexes])

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

    // TODO(nsc) - add a feature flag to enable nerd controls
    // https://github.com/sourcegraph/sourcegraph/pull/49128/files#diff-04df7090c83826679f92f4ee2881b626422057a8e6b59750937e2888d74e411cL152
    const forNerds = false

    return forNerds ? (
        <Menu>
            <>
                <MenuButton
                    className={classNames('text-decoration-none', styles.braindot, dotStyle)}
                    aria-label="Code graph"
                >
                    <Icon aria-hidden={true} svgPath={mdiBrain} />
                </MenuButton>

                <MenuList position={Position.bottomEnd} className={styles.dropdownMenu}>
                    <MenuHeader>
                        Nerd controls
                        <span className="float-right">
                            <Tooltip content="View code intelligence summary">
                                <Link to={`/${repoName}/-/code-graph/dashboard`}>
                                    <Icon aria-hidden={true} svgPath={mdiArrowRightThin} />
                                </Link>
                            </Tooltip>
                        </span>
                    </MenuHeader>

                    <MenuDivider />

                    {/* TODO - add content */}
                    <LoadingSpinner className="mx-2" />
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

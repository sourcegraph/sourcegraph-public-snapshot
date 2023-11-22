import React, { useCallback } from 'react'

import { mdiFolderOpenOutline, mdiFolderOutline, mdiWrench } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Link, Text, Tree, Badge } from '@sourcegraph/wildcard'

import type { CodeIntelIndexerFields, PreciseIndexFields } from '../../../../../graphql-operations'
import { INDEX_COMPLETED_STATES, INDEX_FAILURE_STATES } from '../../constants'
import type { FilterState } from '../../pages/RepoDashboardPage'
import { IndexStateBadge } from '../IndexStateBadge'

import { byKey, getIndexerKey, groupBy, sanitizePath, getIndexRoot, buildTreeData, descendentNames } from './util'

import styles from './DashboardTree.module.scss'

export interface SuggestedIndexWithRoot extends CodeIntelIndexerFields {
    root: string
    comparisonKey: string
}

interface DashboardTreeProps {
    indexes: PreciseIndexFields[]
    suggestedIndexers: SuggestedIndexWithRoot[]
    filter: FilterState
}

export const DashboardTree: React.FunctionComponent<DashboardTreeProps> = ({ indexes, suggestedIndexers, filter }) => {
    const shouldDisplayIndex = useCallback(
        (index: PreciseIndexFields): boolean =>
            // Valid show filter
            (filter.show === 'all' || filter.show === 'indexes') &&
            // Valid language filter
            (filter.language === 'all' || filter.language === getIndexerKey(index)) &&
            // Valid indexState filter
            (filter.indexState === 'all' ||
                (filter.indexState === 'error' && INDEX_FAILURE_STATES.has(index.state)) ||
                (filter.indexState === 'success' && INDEX_COMPLETED_STATES.has(index.state))),
        [filter]
    )

    const shouldDisplayIndexerSuggestion = useCallback(
        (indexer: CodeIntelIndexerFields): boolean =>
            // Valid show filter
            (filter.show === 'all' || filter.show === 'suggestions') &&
            // Valid language filter
            (filter.language === 'all' || filter.language === indexer.key),
        [filter]
    )

    const filteredIndexes = indexes.filter(shouldDisplayIndex)
    const filteredSuggestedIndexers = suggestedIndexers.filter(shouldDisplayIndexerSuggestion)
    const filteredRoots = new Set([
        ...filteredIndexes.map(getIndexRoot),
        ...filteredSuggestedIndexers.map(indexer => indexer.root),
    ])

    const filteredTreeData = buildTreeData(filteredRoots)

    // We always have the root node
    if (filteredTreeData.length === 1) {
        return <Text className="text-muted">No data to display.</Text>
    }

    return (
        <Tree
            data={filteredTreeData}
            defaultExpandedIds={filteredTreeData.map(element => element.id)}
            propagateCollapse={true}
            renderNode={({ element: { id, name: treeRoot, displayName }, handleExpand, ...props }) => {
                const descendentRoots = new Set(descendentNames(filteredTreeData, id).map(sanitizePath))
                const filteredIndexesForRoot = filteredIndexes.filter(index => getIndexRoot(index) === treeRoot)
                const filteredIndexesForDescendents = filteredIndexes.filter(index =>
                    descendentRoots.has(getIndexRoot(index))
                )
                const filteredSuggestedIndexersForRoot = filteredSuggestedIndexers.filter(
                    ({ root }) => sanitizePath(root) === treeRoot
                )
                const filteredSuggestedIndexersForDescendents = filteredSuggestedIndexers.filter(({ root }) =>
                    descendentRoots.has(sanitizePath(root))
                )

                return (
                    <TreeNode
                        onClick={handleExpand}
                        displayName={displayName}
                        indexesByIndexerNameForRoot={groupBy(filteredIndexesForRoot, getIndexerKey)}
                        availableIndexersForRoot={filteredSuggestedIndexersForRoot}
                        numDescendentSuccess={
                            // TODO: Check matches summary num
                            filteredIndexesForDescendents.filter(index => INDEX_COMPLETED_STATES.has(index.state))
                                .length
                        }
                        numDescendentErrors={
                            filteredIndexesForDescendents.filter(index => INDEX_FAILURE_STATES.has(index.state)).length
                        }
                        numDescendentConfigurable={
                            filter.show === 'all' || filter.show === 'suggestions'
                                ? filteredSuggestedIndexersForDescendents.length
                                : 0
                        }
                        {...props}
                    />
                )
            }}
        />
    )
}

interface TreeNodeProps {
    displayName: string
    isBranch: boolean
    isExpanded: boolean
    indexesByIndexerNameForRoot: Map<string, PreciseIndexFields[]>
    availableIndexersForRoot: SuggestedIndexWithRoot[]
    numDescendentErrors: number
    numDescendentSuccess: number
    numDescendentConfigurable: number
    onClick: (event: React.MouseEvent) => {}
}

const TreeNode: React.FunctionComponent<TreeNodeProps> = ({
    displayName,
    isBranch,
    isExpanded,
    indexesByIndexerNameForRoot,
    availableIndexersForRoot,
    numDescendentErrors,
    numDescendentSuccess,
    numDescendentConfigurable,
    onClick,
}) => (
    // We already handle accessibility events for expansion in the <TreeView />
    // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
    <div className={styles.treeNode} onClick={onClick}>
        <div className={classNames('d-inline', !isBranch ? styles.spacer : '')}>
            <Icon
                svgPath={isBranch && isExpanded ? mdiFolderOpenOutline : mdiFolderOutline}
                className={classNames('mr-1', styles.icon)}
                aria-hidden={true}
            />
            {displayName}
            {isBranch &&
                !isExpanded &&
                (numDescendentConfigurable > 0 || numDescendentErrors > 0 || numDescendentSuccess > 0) && (
                    <>
                        {numDescendentConfigurable > 0 && (
                            <Badge variant="primary" className="ml-2" pill={true} small={true}>
                                {numDescendentConfigurable}
                            </Badge>
                        )}
                        {numDescendentErrors > 0 && (
                            <Badge variant="danger" className="ml-2" pill={true} small={true}>
                                {numDescendentErrors}
                            </Badge>
                        )}
                        {numDescendentSuccess > 0 && (
                            <Badge variant="success" className="ml-2" pill={true} small={true}>
                                {numDescendentSuccess}
                            </Badge>
                        )}
                    </>
                )}
        </div>

        <div className="d-flex align-items-center">
            {[...indexesByIndexerNameForRoot?.entries()].sort(byKey).map(([indexerName, indexes]) => (
                <IndexStateBadge
                    key={indexerName}
                    indexes={indexes}
                    className={classNames('text-muted', styles.badge)}
                />
            ))}

            {availableIndexersForRoot.map(indexer => (
                <Badge
                    as={Link}
                    to={`../index-configuration?tab=form#${indexer.comparisonKey}`}
                    variant="outlineSecondary"
                    key={indexer.key}
                    className={classNames('text-muted', styles.badge)}
                >
                    <Icon svgPath={mdiWrench} aria-hidden={true} className="mr-1 text-primary" />
                    Configure {indexer.key}
                </Badge>
            ))}
        </div>
    </div>
)

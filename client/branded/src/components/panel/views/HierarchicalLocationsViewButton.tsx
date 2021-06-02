import classNames from 'classnames'
import React from 'react'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface HierarchicalLocationsViewButtonProps {
    groupKey: string
    groupCount: number
    isActive: boolean
    onClick: (event: React.MouseEvent<HTMLElement>) => void
}

export const HierarchicalLocationsViewButton: React.FunctionComponent<HierarchicalLocationsViewButtonProps> = props => {
    const { groupKey, groupCount, isActive, onClick } = props

    const [isRedesignEnabled] = useRedesignToggle()

    return (
        <button
            type="button"
            className={classNames(
                'list-group-item list-group-item-action hierarchical-locations-view__item',
                isActive && 'active'
            )}
            onClick={onClick}
        >
            <span className="hierarchical-locations-view__item-name" title={groupKey}>
                <span className="hierarchical-locations-view__item-name-text">
                    <RepoLink to={null} repoName={groupKey} />
                </span>
            </span>
            <span
                className={classNames(
                    'hierarchical-locations-view__item-badge',
                    !isRedesignEnabled && 'badge badge-secondary badge-pill ',
                    isRedesignEnabled && 'ml-1'
                )}
            >
                {groupCount}
            </span>
        </button>
    )
}

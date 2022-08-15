import React from 'react'

import classNames from 'classnames'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'

import styles from './HierarchicalLocationsViewButton.module.scss'

interface HierarchicalLocationsViewButtonProps {
    groupKey: string
    groupName: string
    groupCount: number
    isActive: boolean
    onClick: (event: React.MouseEvent<HTMLElement>) => void
}

export const HierarchicalLocationsViewButton: React.FunctionComponent<
    React.PropsWithChildren<HierarchicalLocationsViewButtonProps>
> = props => {
    const { groupKey, groupName, groupCount, isActive, onClick } = props

    return (
        <button
            type="button"
            data-testid="hierarchical-locations-view-button"
            aria-label={`${groupName === 'repo' ? 'Repository' : 'File path'} ${groupKey}`}
            className={classNames(
                'list-group-item list-group-item-action',
                styles.locationButton,
                isActive && 'active'
            )}
            onClick={onClick}
        >
            <span className={styles.locationName} title={groupKey}>
                <span className={styles.locationNameText}>
                    <RepoLink to={null} repoName={groupKey} />
                </span>
            </span>
            <span className={classNames(styles.locationBadge, 'ml-1')}>{groupCount}</span>
        </button>
    )
}

import React from 'react'

import { useLocation } from 'react-router'

import { Link } from '@sourcegraph/wildcard'

import styles from './EmptyDraftChangesetListElement.module.scss'

export const EmptyDraftChangesetListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => {
    const location = useLocation()
    return (
        <div className={styles.emptyDraftChangesetListElementBody}>
            <h3>No changesets exist</h3>
            <div className={styles.emptyDraftChangesetListElementContent}>
                <p className="mt-2">
                    This batch change is a draft. A batch spec must be executed to create changesets.
                </p>
                <Link to={`${location.pathname}/edit`}>Edit the most recent spec.</Link>
            </div>
        </div>
    )
}

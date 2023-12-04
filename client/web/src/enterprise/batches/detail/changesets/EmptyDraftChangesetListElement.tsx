import React from 'react'

import { useLocation } from 'react-router-dom'

import { Link, H3, Text } from '@sourcegraph/wildcard'

import styles from './EmptyDraftChangesetListElement.module.scss'

export const EmptyDraftChangesetListElement: React.FunctionComponent<React.PropsWithChildren<{}>> = () => {
    const location = useLocation()
    return (
        <div className={styles.emptyDraftChangesetListElementBody}>
            <H3>No changesets exist</H3>
            <div className={styles.emptyDraftChangesetListElementContent}>
                <Text className="mt-2">
                    This batch change is a draft. A batch spec must be executed and applied to create changesets.
                </Text>
                <Link to={`${location.pathname}/edit`}>Edit the most recent spec.</Link>
            </div>
        </div>
    )
}

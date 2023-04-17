import React, { useMemo } from 'react'

import classNames from 'classnames'
import { sortBy } from 'lodash'

import { Badge } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

export const RepoMetadata: React.FunctionComponent<{
    keyValuePairs: [string, string | null | undefined][]
    className?: string
    small?: boolean
}> = props => {
    const sortedPairs = useMemo(() => sortBy(props.keyValuePairs), [props.keyValuePairs])
    if (sortedPairs.length === 0) {
        return null
    }
    return (
        <div className={classNames(styles.repoMetadata, props.className, 'd-flex align-items-center flex-wrap')}>
            {sortedPairs.map(([key, value]) => (
                <span className="d-flex align-items-center justify-content-center" key={`${key}:${value}`}>
                    <Badge
                        small={props.small}
                        variant="info"
                        className={classNames({ [styles.repoMetadataKey]: value })}
                    >
                        {key}
                    </Badge>
                    {value && (
                        <Badge small={props.small} variant="secondary" className={styles.repoMetadataValue}>
                            {value}
                        </Badge>
                    )}
                </span>
            ))}
        </div>
    )
}

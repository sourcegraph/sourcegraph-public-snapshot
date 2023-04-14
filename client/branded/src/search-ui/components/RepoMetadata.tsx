import React from 'react'

import classNames from 'classnames'

import { Badge } from '@sourcegraph/wildcard'

import styles from './RepoMetadata.module.scss'

export const RepoMetadata: React.FunctionComponent<{
    keyValuePairs: [string, string | null | undefined][]
    className?: string
    small?: boolean
}> = ({ keyValuePairs, className, small }) => {
    if (keyValuePairs.length === 0) {
        return null
    }
    return (
        <div className={classNames(styles.repoMetadata, className, 'd-flex align-items-center flex-wrap')}>
            {keyValuePairs.map(([key, value]) => (
                <span className="d-flex align-items-center justify-content-center" key={`${key}:${value}`}>
                    <Badge small={small} variant="info" className={classNames({ [styles.repoMetadataKey]: value })}>
                        {key}
                    </Badge>
                    {value && (
                        <Badge small={small} variant="secondary" className={styles.repoMetadataValue}>
                            {value}
                        </Badge>
                    )}
                </span>
            ))}
        </div>
    )
}

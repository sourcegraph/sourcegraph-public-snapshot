import * as React from 'react'

import classNames from 'classnames'

import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { Link } from '@sourcegraph/wildcard'

import styles from './RepoFileSymbolLink.module.scss'

interface RepoFileSymbolLinkProps {
    repoName: string
    repoURL: string
    filePath: string
    fileURL: string
    symbolName?: string
    symbolURL?: string
}

export const RepoFileSymbolLink: React.FunctionComponent<React.PropsWithChildren<RepoFileSymbolLinkProps>> = ({
    repoName,
    repoURL,
    filePath,
    fileURL,
    symbolName,
    symbolURL,
}) => {
    const [fileBase, fileName] = splitPath(filePath)
    return (
        <div className={styles.root}>
            <Link to={repoURL} className={styles.xsHide}>
                {displayRepoName(repoName)}
            </Link>
            <span className={classNames('mx-1', styles.xsHide)}>›</span>
            <Link to={fileURL} className={classNames(symbolURL && styles.xsHide)}>
                <span className={styles.xsHide}>{fileBase ? `${fileBase}/` : null}</span>
                {!symbolURL ? <strong>{fileName}</strong> : <>{fileName}</>}
            </Link>
            {symbolURL && (
                <>
                    <span className={classNames('mx-1', styles.xsHide)}>›</span>
                    <Link to={symbolURL} data-testid="selected-symbol-name">
                        <strong>{symbolName}</strong>
                    </Link>
                </>
            )}
        </div>
    )
}

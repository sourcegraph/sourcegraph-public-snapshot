/* eslint-disable jsx-a11y/no-static-element-interactions */
/* eslint-disable jsx-a11y/click-events-have-key-events */
import React, { useState } from 'react'

import { mdiChevronDown, mdiChevronRight } from '@mdi/js'

import styles from './ContextFiles.module.css'

export interface FileLinkProps {
    path: string
}

export const ContextFiles: React.FunctionComponent<{
    contextFiles: string[]
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
}> = ({ contextFiles, fileLinkComponent: FileLink }) => {
    const [isExpanded, setIsExpanded] = useState(false)

    if (contextFiles.length === 1) {
        return (
            <p>
                Cody read{' '}
                <span className={styles.contextFile}>
                    <FileLink path={contextFiles[0]} />
                </span>{' '}
                to provide an answer.
            </p>
        )
    }

    if (isExpanded) {
        return (
            <div className={styles.contextFilesExpanded}>
                <span className={styles.contextFilesToggleIcon} onClick={() => setIsExpanded(false)}>
                    <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                </span>
                <div>
                    <div className={styles.contextFilesListTitle} onClick={() => setIsExpanded(false)}>
                        Cody read the following files to provide an answer:
                    </div>
                    <ul className={styles.contextFilesListContainer}>
                        {contextFiles.map(file => (
                            <li key={file} className={styles.contextFile}>
                                <FileLink path={file} />
                            </li>
                        ))}
                    </ul>
                </div>
            </div>
        )
    }

    return (
        <div className={styles.contextFilesCollapsed} onClick={() => setIsExpanded(true)}>
            <span className={styles.contextFilesToggleIcon}>
                <Icon aria-hidden={true} svgPath={mdiChevronRight} />
            </span>
            <div className={styles.contextFilesCollapsedText}>
                <span>
                    Cody read <span className={styles.contextFile}>{contextFiles[0].split('/').pop()}</span> and{' '}
                    {contextFiles.length - 1} other {contextFiles.length > 2 ? 'files' : 'file'} to provide an answer.
                </span>
            </div>
        </div>
    )
}

const Icon: React.FC<{ svgPath: string }> = ({ svgPath }) => (
    <svg role="img" height={24} width={24} viewBox="0 0 24 24" fill="currentColor">
        <path d={svgPath} />
    </svg>
)

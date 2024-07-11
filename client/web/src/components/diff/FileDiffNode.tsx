import React, { useState, useCallback } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'
import prettyBytes from 'pretty-bytes'
import { useLocation } from 'react-router-dom'

import { dirname } from '@sourcegraph/common'
import { Button, Badge, Link, Icon, Text, createLinkUrl, Tooltip } from '@sourcegraph/wildcard'

import type { FileDiffFields } from '../../graphql-operations'
import type { DiffMode } from '../../repo/commit/RepositoryCommitPage'
import { isPerforceChangelistMappingEnabled } from '../../repo/utils'

import { DiffStat, DiffStatSquares } from './DiffStat'
import { FileDiffHunks } from './FileDiffHunks'

import styles from './FileDiffNode.module.scss'

export interface FileDiffNodeProps {
    node: FileDiffFields
    lineNumbers: boolean
    className?: string

    /** Reflect selected line in url */
    persistLines?: boolean
    diffMode?: DiffMode
}

/** A file diff. */
export const FileDiffNode: React.FunctionComponent<React.PropsWithChildren<FileDiffNodeProps>> = ({
    lineNumbers,
    node,
    className,
    persistLines,
    diffMode = 'unified',
}) => {
    const location = useLocation()
    const [expanded, setExpanded] = useState<boolean>(true)
    const [renderDeleted, setRenderDeleted] = useState<boolean>(false)

    const toggleExpand = useCallback((): void => {
        setExpanded(!expanded)
    }, [expanded])

    const onClickToViewDeleted = useCallback((): void => {
        setRenderDeleted(true)
    }, [])

    let path: React.ReactNode
    if (node.newPath && (node.newPath === node.oldPath || !node.oldPath)) {
        path = <span title={node.newPath}>{node.newPath}</span>
    } else if (node.newPath && node.oldPath && node.newPath !== node.oldPath) {
        path = (
            <span title={`${node.oldPath} → ${node.newPath}`}>
                {node.oldPath} → {node.newPath}
            </span>
        )
    } else {
        // By process of elimination (that TypeScript is unfortunately unable to infer, except
        // by reorganizing this code in a way that's much more complex to humans), node.oldPath
        // is non-null.
        path = <span title={node.oldPath!}>{node.oldPath}</span>
    }

    let stat: React.ReactNode
    // If one of the files was binary, display file size change instead of DiffStat.
    if (node.oldFile?.binary || node.newFile?.binary) {
        const sizeChange = (node.newFile?.byteSize ?? 0) - (node.oldFile?.byteSize ?? 0)
        const className = sizeChange >= 0 ? 'text-success' : 'text-danger'
        stat = <strong className={classNames(className, 'code')}>{prettyBytes(sizeChange)}</strong>
    } else {
        stat = (
            <>
                <DiffStat className="mr-1" {...node.stat} />
                <DiffStatSquares {...node.stat} />
            </>
        )
    }

    const anchor = `diff-${node.internalID}`

    const gitBlobURL =
        isPerforceChangelistMappingEnabled() &&
        node.mostRelevantFile.__typename === 'GitBlob' &&
        node.mostRelevantFile.changelistURL
            ? node.mostRelevantFile.changelistURL
            : node.mostRelevantFile.url

    return (
        <>
            {/* The empty <a> tag is to allow users to anchor links to the top of this file diff node */}
            <Link to="" id={anchor} aria-hidden={true} tabIndex={-1} />
            <li className={classNames('test-file-diff-node', styles.fileDiffNode, className)}>
                <div className={styles.header}>
                    <Button
                        aria-label={expanded ? 'Hide file diff' : 'Show file diff'}
                        variant="icon"
                        className="mr-2"
                        onClick={toggleExpand}
                        size="sm"
                    >
                        <Icon svgPath={expanded ? mdiChevronUp : mdiChevronDown} aria-hidden={true} />
                    </Button>
                    <div className={classNames('align-items-baseline', styles.headerPathStat)}>
                        {!node.oldPath && (
                            <Badge variant="success" className="text-uppercase mr-2">
                                Added
                            </Badge>
                        )}
                        {!node.newPath && (
                            <Badge variant="danger" className="text-uppercase mr-2">
                                Deleted
                            </Badge>
                        )}
                        {node.newPath && node.oldPath && node.newPath !== node.oldPath && (
                            <Badge variant="warning" className="text-uppercase mr-2">
                                {dirname(node.newPath) !== dirname(node.oldPath) ? 'Moved' : 'Renamed'}
                            </Badge>
                        )}
                        {stat}
                        {node.mostRelevantFile.__typename === 'GitBlob' ? (
                            <Tooltip content="View file at revision">
                                <Link to={gitBlobURL} className="mr-0 ml-2 fw-bold">
                                    <strong>{path}</strong>
                                </Link>
                            </Tooltip>
                        ) : (
                            <span className="ml-2">{path}</span>
                        )}
                        <Tooltip content="Pin diff">
                            <Link
                                to={createLinkUrl({ ...location, hash: anchor })}
                                className={classNames('ml-2', styles.headerPath)}
                            >
                                #
                            </Link>
                        </Tooltip>
                    </div>
                </div>
                {expanded &&
                    (node.oldFile?.binary || node.newFile?.binary ? (
                        <div className="text-muted m-2">Binary files can't be rendered.</div>
                    ) : !node.newPath && !renderDeleted ? (
                        <div className="text-muted m-2">
                            <Text className="mb-0">Deleted files aren't rendered by default.</Text>
                            <Button className="m-0 p-0" onClick={onClickToViewDeleted} variant="link">
                                Click here to view.
                            </Button>
                        </div>
                    ) : (
                        <FileDiffHunks
                            className={styles.hunks}
                            fileDiffAnchor={anchor}
                            persistLines={persistLines}
                            hunks={node.hunks}
                            lineNumbers={lineNumbers}
                            diffMode={diffMode}
                        />
                    ))}
            </li>
        </>
    )
}

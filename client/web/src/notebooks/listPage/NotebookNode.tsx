import React, { useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import StarIcon from 'mdi-react/StarIcon'
import StarOutlineIcon from 'mdi-react/StarOutlineIcon'

import { renderMarkdown, pluralize } from '@sourcegraph/common'
import { IMarkdownBlock, NotebookBlock } from '@sourcegraph/shared/src/schema'
import { Link, Badge, Icon } from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/time/Timestamp'
import { NotebookFields } from '../../graphql-operations'
import { PageRoutes } from '../../routes.constants'

import styles from './NotebookNode.module.scss'

export interface NotebookNodeProps {
    node: NotebookFields
    location: H.Location
    history: H.History
}

// Find the first Markdown block in the notebook, and use the first line in the block
// as the notebook description.
function getNotebookDescription(blocks: NotebookBlock[]): string {
    const firstMarkdownBlock = blocks.find<IMarkdownBlock>(
        (block): block is IMarkdownBlock => block.__typename === 'MarkdownBlock'
    )
    if (!firstMarkdownBlock) {
        return ''
    }
    const renderedPlainTextMarkdown = renderMarkdown(firstMarkdownBlock.markdownInput, { plainText: true })
    return renderedPlainTextMarkdown.split('\n')[0]
}

export const NotebookNode: React.FunctionComponent<React.PropsWithChildren<NotebookNodeProps>> = ({
    node,
}: NotebookNodeProps) => {
    const description = useMemo(() => getNotebookDescription(node.blocks), [node.blocks])
    return (
        <li className={classNames('py-3', styles.notebookNode)}>
            <div className="d-flex align-items-center">
                <Link to={PageRoutes.Notebook.replace(':id', node.id)} className={styles.notebookLink}>
                    <strong>{node.title}</strong>
                </Link>
                {!node.public && (
                    <Badge variant="secondary" pill={true} className={classNames('ml-1', styles.privateBadge)} as="div">
                        Private
                    </Badge>
                )}
            </div>
            {description && <div className={classNames('text-muted mt-1', styles.description)}>{description}</div>}
            <div className={classNames('text-muted mt-2 d-flex align-items-center', styles.meta)}>
                <span className="mr-3">
                    Created by {node.creator ? <strong>@{node.creator.username}</strong> : <span>unknown user</span>}
                </span>
                <span className="mr-3">
                    {node.blocks.length} {pluralize('block', node.blocks.length, 'blocks')}
                </span>
                <span className="d-flex align-items-center mr-3">
                    {node.viewerHasStarred ? (
                        <Icon
                            className={classNames(styles.notebookStarIcon, styles.notebookStarIconActive)}
                            as={StarIcon}
                        />
                    ) : (
                        <Icon className={styles.notebookStarIcon} as={StarOutlineIcon} />
                    )}
                    <span className="ml-1">{node.stars.totalCount}</span>
                </span>
                <span className="mr-3">
                    Updated <Timestamp date={node.updatedAt} noAbout={true} />
                </span>
                <span className="mr-3">
                    Created <Timestamp date={node.createdAt} noAbout={true} />
                </span>
            </div>
        </li>
    )
}

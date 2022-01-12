import classNames from 'classnames'
import * as H from 'history'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Badge } from '@sourcegraph/wildcard'

import { NotebookFields } from '../../../graphql-operations'
import { PageRoutes } from '../../../routes.constants'

import styles from './NotebookNode.module.scss'

export interface NotebookNodeProps {
    node: NotebookFields
    location: H.Location
    history: H.History
}

export const NotebookNode: React.FunctionComponent<NotebookNodeProps> = ({ node }: NotebookNodeProps) => (
    <div className={classNames('py-3 d-flex align-items-center', styles.notebookNode)}>
        <div className={classNames('flex-grow-1', styles.left)}>
            <div>
                <Link to={PageRoutes.Notebook.replace(':id', node.id)}>
                    <strong>{node.title}</strong>
                </Link>
                {!node.public && (
                    <Badge variant="secondary" pill={true} className="ml-1" as="div">
                        Private
                    </Badge>
                )}
            </div>

            <div className={classNames('text-muted mt-1', styles.description)}>
                <small>
                    Created by {node.creator ? node.creator.username : 'unknown user'}, {node.blocks.length}{' '}
                    {pluralize('block', node.blocks.length, 'blocks')}
                </small>
            </div>
        </div>
        <div className={classNames('text-muted d-flex', styles.right)}>
            <div>
                Updated <Timestamp date={node.updatedAt} noAbout={true} />, created{' '}
                <Timestamp date={node.createdAt} noAbout={true} />
            </div>
        </div>
    </div>
)

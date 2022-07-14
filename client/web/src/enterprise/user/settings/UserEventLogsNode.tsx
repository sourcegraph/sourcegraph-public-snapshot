import React from 'react'

import classNames from 'classnames'

import { Link, Code } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../components/time/Timestamp'
import { UserEventLogFields } from '../../../graphql-operations'

import styles from './UserEventLogsPage.module.scss'

interface UserEventNodeProps {
    /**
     * The user to display in this list item.
     */
    node: UserEventLogFields
}

export const UserEventNode: React.FunctionComponent<React.PropsWithChildren<UserEventNodeProps>> = ({
    node,
}: UserEventNodeProps) => (
    <li className={classNames('list-group-item', styles.eventLog)}>
        <div className="d-flex align-items-center justify-content-between">
            <Code>{node.name}</Code>
            <div>
                <Timestamp date={node.timestamp} />
            </div>
        </div>
        <div className="text-break">
            <small>
                From: {node.source}{' '}
                {node.url && (
                    <span>
                        (<Link to={node.url}>{node.url}</Link>)
                    </span>
                )}
            </small>
        </div>
    </li>
)

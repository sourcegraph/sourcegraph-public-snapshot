import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { AnalyticsUserActivity } from '../../../graphql-operations'

import styles from './index.module.scss'

interface UserNodeProps {
    node: AnalyticsUserActivity
    setFilterPackage?: (node: AnalyticsUserActivity) => void
}

export const UserNode: React.FunctionComponent<React.PropsWithChildren<UserNodeProps>> = ({ node }) => {
    return (
        <li className="list-group-item px-0 py-2">
            <div className={styles.node}>
                <div className={styles.user}>
                    <Link to={`/users/${node.userID}`}>{node.username}</Link>
                </div>
                {node.periods.map(p => {
                    return (
                        <div className={styles.period} key={node.userID + ':' + p.date}>
                            {p.count}
                        </div>
                    )
                })}
                <div className={styles.period}>{node.totalEventCount}</div>
            </div>
        </li>
    )
}

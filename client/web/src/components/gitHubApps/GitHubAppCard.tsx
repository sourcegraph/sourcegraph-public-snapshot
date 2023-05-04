import React from 'react'

import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Link, H3 } from '@sourcegraph/wildcard'

import styles from './GitHubAppCard.module.scss'

interface GitHubApp {
    id: string
    appID: number
    appURL: string
    name: string
    logo?: string
    createdAt: string
    updatedAt: string
}

interface GitHubAppCardProps {
    app: GitHubApp
    className?: string
}

export const GitHubAppCard: React.FC<GitHubAppCardProps> = ({ app, className = '' }) => (
    <Link
        className={classNames('d-flex align-items-center p-2 text-body text-decoration-none', className ?? '')}
        to={`./${app.id}`}
    >
        <img className={classNames('mr-3', styles.logo)} src={app.logo} alt="app logo" aria-hidden={true} />
        <span>
            <H3 className="mt-1 mb-0">{app.name}</H3>
            <small className="text-muted">AppID: {app.appID}</small>
        </span>
        <span className="ml-auto">
            Created <Timestamp date={app.createdAt} />
        </span>
    </Link>
)

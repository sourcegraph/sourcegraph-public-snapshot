import React, { useCallback, useState } from 'react'

import { mdiDelete } from '@mdi/js'
import classNames from 'classnames'
import { DeleteGitHubAppResult, DeleteGitHubAppVariables } from 'src/graphql-operations'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, Link, H3, Icon, Tooltip } from '@sourcegraph/wildcard'

import { DELETE_GITHUB_APP_BY_ID_QUERY } from './backend'

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

export const GitHubAppCard: React.FC<GitHubAppCardProps> = ({ app, className = '' }) => {
    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)
    const [deleteGitHubApp] = useMutation<DeleteGitHubAppResult, DeleteGitHubAppVariables>(
        DELETE_GITHUB_APP_BY_ID_QUERY
    )

    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!window.confirm(`Delete the GitHub App ${app.name}?`)) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteGitHubApp({
                variables: { gitHubApp: app.id },
            })
            setIsDeleting(false)
        } catch (error) {
            setIsDeleting(asError(error))
        } finally {
            window.location.reload()
        }
    }, [app, deleteGitHubApp])

    return (
        <div className={classNames('d-flex align-items-center p-2 text-body text-decoration-none')}>
            <Link
                className={classNames('d-flex align-items-center p-2 text-body text-decoration-none', className ?? '')}
                to={`./${app.id}`}
            >
                <img className={classNames('mr-3', styles.logo)} src={app.logo} alt="app logo" aria-hidden={true} />
                <span>
                    <H3 className="mt-1 mb-0">{app.name}</H3>
                    <small className="text-muted">AppID: {app.appID}</small>
                </span>
            </Link>
            <span className="ml-auto mr-1">
                Created <Timestamp date={app.createdAt} />
            </span>
            <Tooltip content="Delete GitHub App">
                <Button
                    aria-label="Delete"
                    onClick={onDelete}
                    disabled={isDeleting === true}
                    variant="danger"
                    size="sm"
                >
                    <Icon aria-hidden={true} svgPath={mdiDelete} />
                    {' Delete'}
                </Button>
            </Tooltip>
        </div>
    )
}

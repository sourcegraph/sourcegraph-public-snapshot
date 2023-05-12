import React, { useCallback } from 'react'

import { mdiDelete, mdiOpenInNew, mdiCogOutline } from '@mdi/js'
import classNames from 'classnames'
import { DeleteGitHubAppResult, DeleteGitHubAppVariables } from 'src/graphql-operations'

import { useMutation } from '@sourcegraph/http-client'
import { Button, Link, Icon, Tooltip, Container, AnchorLink, ButtonLink } from '@sourcegraph/wildcard'

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
    refetch: () => void
}

export const GitHubAppCard: React.FC<GitHubAppCardProps> = ({ app, refetch }) => {
    const [deleteGitHubApp, { loading }] = useMutation<DeleteGitHubAppResult, DeleteGitHubAppVariables>(
        DELETE_GITHUB_APP_BY_ID_QUERY
    )

    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!window.confirm(`Delete the GitHub App ${app.name}?`)) {
            return
        }
        try {
            await deleteGitHubApp({
                variables: { gitHubApp: app.id },
            })
        } finally {
            refetch()
        }
    }, [app, deleteGitHubApp, refetch])

    return (
        <Container className="d-flex align-items-center mb-2 p-3">
            <Link className={classNames(styles.appLink, 'd-flex align-items-center text-decoration-none')} to={app.id}>
                <img className={classNames('mr-3', styles.logo)} src={app.logo} alt="app logo" aria-hidden={true} />
                <span>
                    <div className="font-weight-bold">{app.name}</div>
                    <div className="text-muted">AppID: {app.appID}</div>
                </span>
            </Link>
            <div className="ml-auto">
                <AnchorLink to={app.appURL} target="_blank" className="mr-3">
                    <small>
                        View In GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} />
                    </small>
                </AnchorLink>
                <ButtonLink className="mr-2" aria-label="Edit" to={app.id} variant="secondary" size="sm">
                    <Icon aria-hidden={true} svgPath={mdiCogOutline} /> Edit
                </ButtonLink>
                <Tooltip content="Delete GitHub App">
                    <Button aria-label="Delete" onClick={onDelete} disabled={loading} variant="danger" size="sm">
                        <Icon aria-hidden={true} svgPath={mdiDelete} /> Delete
                    </Button>
                </Tooltip>
            </div>
        </Container>
    )
}

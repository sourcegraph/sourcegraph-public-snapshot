import React, { useState } from 'react'

import { mdiDelete, mdiOpenInNew, mdiCogOutline } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, Container, AnchorLink, ButtonLink } from '@sourcegraph/wildcard'

import { AppLogo } from './AppLogo'
import { RemoveGitHubAppModal } from './RemoveGitHubAppModal'

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
    const [removeModalOpen, setRemoveModalOpen] = useState<boolean>(false)

    return (
        <Container className="d-flex align-items-center mb-2 p-3">
            {removeModalOpen && (
                <RemoveGitHubAppModal onCancel={() => setRemoveModalOpen(false)} afterDelete={refetch} app={app} />
            )}
            <span className={classNames(styles.appLink, 'd-flex align-items-center')}>
                <AppLogo src={app.logo} name={app.name} className={classNames(styles.logo, 'mr-2')} />
                <span>
                    <div className="font-weight-bold">{app.name}</div>
                    <div className="text-muted">AppID: {app.appID}</div>
                </span>
            </span>
            <div className="ml-auto">
                <AnchorLink to={app.appURL} target="_blank" className="mr-3">
                    <small>
                        View In GitHub <Icon inline={true} svgPath={mdiOpenInNew} aria-hidden={true} />
                    </small>
                </AnchorLink>
                <ButtonLink className="mr-2" aria-label="Edit" to={app.id} variant="secondary" size="sm">
                    <Icon aria-hidden={true} svgPath={mdiCogOutline} /> Edit
                </ButtonLink>
                <Button
                    aria-label="Remove GitHub App"
                    onClick={() => setRemoveModalOpen(true)}
                    variant="danger"
                    size="sm"
                >
                    <Icon aria-hidden={true} svgPath={mdiDelete} /> Remove
                </Button>
            </div>
        </Container>
    )
}

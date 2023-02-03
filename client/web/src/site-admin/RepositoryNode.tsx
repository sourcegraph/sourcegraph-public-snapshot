import React from 'react'

import { mdiCloudDownload, mdiCog, mdiFileDocumentOutline } from '@mdi/js'
import classNames from 'classnames'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { Alert, Badge, Button, Icon, Link, LoadingSpinner, Tooltip, LinkOrSpan } from '@sourcegraph/wildcard'

import { SiteAdminRepositoryFields } from '../graphql-operations'

import { ExternalRepositoryIcon } from './components/ExternalRepositoryIcon'
import { RepoMirrorInfo } from './components/RepoMirrorInfo'

import styles from './RepositoryNode.module.scss'

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

export const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => {
    let status = 'queued'
    if (node.mirrorInfo.cloned && !node.mirrorInfo.lastError) {
        status = 'cloned'
    } else if (node.mirrorInfo.cloneInProgress) {
        status = 'cloning'
    } else if (node.mirrorInfo.lastError) {
        status = 'failed'
    }

    return (
        <li
            className="repository-node list-group-item px-0 py-2"
            data-test-repository={node.name}
            data-test-cloned={node.mirrorInfo.cloned}
        >
            <div className="d-flex align-items-center justify-content-between">
                <div className="d-flex col-7 pl-0">
                    <div className={classNames('d-flex col-2 px-0 justify-content-between h-100', styles.badgeWrapper)}>
                        <Badge
                            className={classNames(
                                styles[status as keyof typeof styles],
                                'py-0 px-1 text-uppercase font-weight-normal'
                            )}
                        >
                            {status}
                        </Badge>
                        {node.mirrorInfo.cloneInProgress && <LoadingSpinner />}
                    </div>

                    <div className="d-flex flex-column ml-2">
                        <div>
                            <ExternalRepositoryIcon externalRepo={node.externalRepository} />
                            <RepoLink repoName={node.name} to={node.url} />
                        </div>
                        <RepoMirrorInfo mirrorInfo={node.mirrorInfo} />
                    </div>
                </div>

                <div className="col-auto pr-0">
                    {/* TODO: Enable 'CLONE NOW' to enqueue the repo
                    {!node.mirrorInfo.cloned && !node.mirrorInfo.cloneInProgress && !node.mirrorInfo.lastError && (
                        <Button to={node.url} variant="secondary" size="sm" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiCloudDownload} /> Clone now
                        </Button>
                    )}{' '} */}
                    {node.mirrorInfo.cloned && !node.mirrorInfo.lastError && !node.mirrorInfo.cloneInProgress && (
                        <Tooltip content="Repository settings">
                            <Button to={`/${node.name}/-/settings`} variant="secondary" size="sm" as={Link}>
                                <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
                            </Button>
                        </Tooltip>
                    )}
                    {node.mirrorInfo.lastError && (
                        <Button to={`/${node.name}/-/settings/mirror`} variant="secondary" size="sm" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiFileDocumentOutline} /> See errors
                        </Button>
                    )}
                </div>
            </div>

            {node.mirrorInfo.isCorrupted && (
                <div className={styles.alertWrapper}>
                    <Alert variant="danger">
                        Repository is corrupt.{' '}
                        <LinkOrSpan to={`/${node.name}/-/settings/mirror`}>More details</LinkOrSpan>
                    </Alert>
                </div>
            )}
        </li>
    )
}

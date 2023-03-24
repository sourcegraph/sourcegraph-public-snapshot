import React, { useState } from 'react'

import { mdiBrain, mdiClose, mdiCog, mdiFileDocumentOutline } from '@mdi/js'
import classNames from 'classnames'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import {
    Alert,
    Badge,
    Button,
    H4,
    Icon,
    Link,
    LinkOrSpan,
    LoadingSpinner,
    MenuDivider,
    Popover,
    PopoverContent,
    PopoverTail,
    PopoverTrigger,
    Position,
    Tooltip,
} from '@sourcegraph/wildcard'

import { SiteAdminRepositoryFields } from '../graphql-operations'

import { ExternalRepositoryIcon } from './components/ExternalRepositoryIcon'
import { RepoMirrorInfo } from './components/RepoMirrorInfo'

import styles from './RepositoryNode.module.scss'

const RepositoryStatusBadge: React.FunctionComponent<{ status: string }> = ({ status }) => (
    <Badge className={classNames(styles[status as keyof typeof styles], 'py-0 px-1 text-uppercase font-weight-normal')}>
        {status}
    </Badge>
)

const parseRepositoryStatus = (repository: SiteAdminRepositoryFields): string => {
    let status = 'queued'
    if (repository.mirrorInfo.cloned && !repository.mirrorInfo.lastError) {
        status = 'cloned'
    } else if (repository.mirrorInfo.cloneInProgress) {
        status = 'cloning'
    } else if (repository.mirrorInfo.lastError) {
        status = 'failed'
    }
    return status
}

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

export const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)

    return (
        <li
            className="repository-node list-group-item px-0 py-2"
            data-test-repository={node.name}
            data-test-cloned={node.mirrorInfo.cloned}
        >
            <div className="d-flex align-items-md-center justify-content-md-between flex-column flex-md-row">
                <div
                    className={classNames(
                        'd-flex col-4 col-md-auto align-items-center pl-0 mt-2 mt-md-0',
                        styles.repoDescription
                    )}
                >
                    <ExternalRepositoryIcon externalRepo={node.externalRepository} />
                    <RepoLink repoName={node.name} to={node.url} />
                </div>

                <div className="d-flex justify-content-start col-4 col-md-2 pl-0 ml-md-2 mt-2 mt-md-0">
                    <RepositoryStatusBadge status={parseRepositoryStatus(node)} />
                    {node.mirrorInfo.cloneInProgress && <LoadingSpinner className="ml-2" />}
                </div>

                <div className="col-md-4 pr-0 pl-0 mt-2 mt-md-0 d-flex align-items-start justify-content-start justify-content-md-end">
                    {/* TODO: Enable 'CLONE NOW' to enqueue the repo
                    {!node.mirrorInfo.cloned && !node.mirrorInfo.cloneInProgress && !node.mirrorInfo.lastError && (
                        <Button to={node.url} variant="secondary" size="sm" as={Link}>
                            <Icon aria-hidden={true} svgPath={mdiCloudDownload} /> Clone now
                        </Button>
                    )}{' '} */}
                    {!window.location.pathname.includes('/setup') && (
                        <Tooltip content="Repository code graph data">
                            <Button
                                to={`/${node.name}/-/code-graph`}
                                variant="secondary"
                                size="sm"
                                className="mr-1"
                                as={Link}
                            >
                                <Icon aria-hidden={true} svgPath={mdiBrain} /> Code graph data
                            </Button>
                        </Tooltip>
                    )}
                    {node.mirrorInfo.cloned && !node.mirrorInfo.lastError && !node.mirrorInfo.cloneInProgress && (
                        <Tooltip content="Repository settings">
                            <Button to={`/${node.name}/-/settings`} variant="secondary" size="sm" as={Link}>
                                <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
                            </Button>
                        </Tooltip>
                    )}
                    {node.mirrorInfo.lastError && (
                        <Popover isOpen={isPopoverOpen} onOpenChange={event => setIsPopoverOpen(event.isOpen)}>
                            <PopoverTrigger as={Button} variant="secondary" size="sm" aria-label="See errors">
                                <Icon aria-hidden={true} svgPath={mdiFileDocumentOutline} /> See errors
                            </PopoverTrigger>

                            <PopoverContent position={Position.left} className={styles.errorContent}>
                                <div className="d-flex">
                                    <H4 className="m-2">
                                        <RepositoryStatusBadge status={parseRepositoryStatus(node)} />
                                        <ExternalRepositoryIcon
                                            externalRepo={node.externalRepository}
                                            className="mx-2"
                                        />
                                        <RepoLink repoName={node.name} to={null} />
                                    </H4>

                                    <Button
                                        aria-label="Dismiss error"
                                        variant="icon"
                                        className="ml-auto mr-2"
                                        onClick={() => setIsPopoverOpen(false)}
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiClose} />
                                    </Button>
                                </div>

                                <MenuDivider />

                                <Alert variant="warning" className={classNames('m-2', styles.alertOverflow)}>
                                    <H4>Error syncing repository:</H4>
                                    {node.mirrorInfo.lastError}
                                </Alert>
                            </PopoverContent>
                            <PopoverTail size="sm" />
                        </Popover>
                    )}
                </div>
            </div>

            <div className="w-100">
                <RepoMirrorInfo mirrorInfo={node.mirrorInfo} />
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

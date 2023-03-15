import React, { useState } from 'react'

import { mdiCog, mdiClose, mdiFileDocumentOutline, mdiBrain, mdiDotsHorizontal } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import {
    Alert,
    Badge,
    Button,
    Icon,
    H4,
    LoadingSpinner,
    LinkOrSpan,
    PopoverTrigger,
    PopoverContent,
    PopoverTail,
    Popover,
    Position,
    MenuDivider,
    MenuButton,
    MenuList,
    MenuItem,
    Menu,
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
    const navigate = useNavigate()
    return (
        <li
            className="repository-node list-group-item px-0 py-2"
            data-test-repository={node.name}
            data-test-cloned={node.mirrorInfo.cloned}
        >
            <div className="d-flex align-items-center justify-content-between">
                <div className="d-flex col-10 pl-0">
                    <div className={classNames('col-2 px-0 my-auto h-100', styles.badgeWrapper)}>
                        <RepositoryStatusBadge status={parseRepositoryStatus(node)} />
                        {node.mirrorInfo.cloneInProgress && <LoadingSpinner className="ml-2" />}
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
                    <Menu>
                        <MenuButton outline={true} aria-label="Repository action">
                            <Icon svgPath={mdiDotsHorizontal} inline={false} aria-hidden={true} />
                        </MenuButton>
                        <MenuList position={Position.bottomEnd}>
                            <MenuItem
                                as={Button}
                                disabled={window.location.pathname.includes('/setup')}
                                onSelect={() => navigate(`/${node.name}/-/code-graph`)}
                                className="p-2"
                            >
                                <Icon aria-hidden={true} svgPath={mdiBrain} className="mr-1" />
                                Code graph data
                            </MenuItem>
                            <MenuItem
                                as={Button}
                                disabled={
                                    !(
                                        node.mirrorInfo.cloned &&
                                        !node.mirrorInfo.lastError &&
                                        !node.mirrorInfo.cloneInProgress
                                    )
                                }
                                onSelect={() => navigate(`/${node.name}/-/settings`)}
                                className="p-2"
                            >
                                <Icon aria-hidden={true} svgPath={mdiCog} className="mr-1" />
                                Settings
                            </MenuItem>
                        </MenuList>
                    </Menu>
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

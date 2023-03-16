import React, { useState } from 'react'

import {
    mdiCog,
    mdiClose,
    mdiFileDocumentOutline,
    mdiBrain,
    mdiDotsVertical,
    mdiDatabaseRefresh,
    mdiRefresh,
} from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useMutation, useQuery } from '@sourcegraph/http-client'
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

import {
    RecloneRepositoryResult,
    RecloneRepositoryVariables,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
    SiteAdminRepositoryFields,
    UpdateMirrorRepositoryResult,
    UpdateMirrorRepositoryVariables,
} from '../graphql-operations'
import { FETCH_SETTINGS_AREA_REPOSITORY_GQL } from '../repo/settings/backend'

import { RECLONE_REPOSITORY_MUTATION, UPDATE_MIRROR_REPOSITORY } from './backend'
import { ExternalRepositoryIcon } from './components/ExternalRepositoryIcon'
import { RepoMirrorInfo } from './components/RepoMirrorInfo'

import styles from './RepositoryNode.module.scss'

const RepositoryStatusBadge: React.FunctionComponent<{ status: string }> = ({ status }) => (
    <Badge className={classNames(styles[status as keyof typeof styles], 'py-0 px-1 text-uppercase font-weight-normal')}>
        {status}
    </Badge>
)

const parseRepositoryStatus = (repo: SiteAdminRepositoryFields): string => {
    let status = 'queued'
    if (repo.mirrorInfo.cloned && !repo.mirrorInfo.lastError) {
        status = 'cloned'
    } else if (repo.mirrorInfo.cloneInProgress) {
        status = 'cloning'
    } else if (repo.mirrorInfo.lastError) {
        status = 'failed'
    }
    return status
}

const repoClonedAndHealthy = (repo: SiteAdminRepositoryFields): boolean =>
    repo.mirrorInfo.cloned && !repo.mirrorInfo.lastError && !repo.mirrorInfo.cloneInProgress

interface RepositoryNodeProps {
    node: SiteAdminRepositoryFields
}

const updateNodeFromData = (node: SiteAdminRepositoryFields, data: SettingsAreaRepositoryResult | undefined): void => {
    if (data?.repository && data.repository?.mirrorInfo) {
        node.mirrorInfo.lastError = data.repository.mirrorInfo.lastError
        node.mirrorInfo.cloned = data.repository.mirrorInfo.cloned
        node.mirrorInfo.cloneInProgress = data.repository.mirrorInfo.cloneInProgress
        node.mirrorInfo.updatedAt = data.repository.mirrorInfo.updatedAt
        node.mirrorInfo.isCorrupted = data.repository.mirrorInfo.isCorrupted
        node.mirrorInfo.corruptionLogs = data.repository.mirrorInfo.corruptionLogs
    }
}

export const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({ node }) => {
    const [isPopoverOpen, setIsPopoverOpen] = useState(false)
    const navigate = useNavigate()
    const [recloneRepository] = useMutation<RecloneRepositoryResult, RecloneRepositoryVariables>(
        RECLONE_REPOSITORY_MUTATION,
        {
            variables: { repo: node.id },
        }
    )
    const [updateRepo] = useMutation<UpdateMirrorRepositoryResult, UpdateMirrorRepositoryVariables>(
        UPDATE_MIRROR_REPOSITORY,
        { variables: { repository: node.id } }
    )
    const { data, refetch } = useQuery<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        {
            variables: { name: node.name },
            pollInterval: 3000,
        }
    )
    const recloneAndFetch = async (): Promise<void> => {
        await recloneRepository()
        await refetch()
        updateNodeFromData(node, data)
    }

    return (
        <li
            className="repository-node list-group-item px-0 py-2"
            data-test-repository={node.name}
            data-test-cloned={node.mirrorInfo.cloned}
        >
            <div className="d-flex align-items-center justify-content-between">
                <div className="d-flex col-11 pl-0">
                    <div className={classNames('col-1 px-0 my-auto h-100', styles.badgeWrapper)}>
                        <RepositoryStatusBadge status={parseRepositoryStatus(node)} />
                        {node.mirrorInfo.cloneInProgress && <LoadingSpinner className="ml-2" />}
                    </div>

                    <div className="col-9 d-flex flex-column ml-2">
                        <div>
                            <ExternalRepositoryIcon externalRepo={node.externalRepository} />
                            <RepoLink repoName={node.name} to={node.url} />
                        </div>
                        <RepoMirrorInfo mirrorInfo={node.mirrorInfo} />
                    </div>
                    <div className="col-2 d-flex align-items-center">
                        {node.name === 'dev.azure.com/sgtestazure/sgtestazure/mytest' && (
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
                <div className="col-auto">
                    {!window.location.pathname.includes('/setup') && (
                        <Menu>
                            <MenuButton outline={true} aria-label="Repository action">
                                <Icon svgPath={mdiDotsVertical} inline={false} aria-hidden={true} />
                            </MenuButton>
                            <MenuList position={Position.bottomEnd}>
                                <MenuItem
                                    as={Button}
                                    disabled={!repoClonedAndHealthy(node)}
                                    onSelect={() => navigate(`/${node.name}/-/code-graph`)}
                                    className="p-2"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiBrain} className="mr-1" />
                                    Code graph data
                                </MenuItem>
                                <MenuItem
                                    as={Button}
                                    disabled={!repoClonedAndHealthy(node)}
                                    onSelect={() => updateRepo()}
                                    className="p-2"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiRefresh} className="mr-1" />
                                    Refetch
                                </MenuItem>
                                <MenuItem
                                    as={Button}
                                    disabled={!repoClonedAndHealthy(node)}
                                    onSelect={() => recloneAndFetch()}
                                    className="p-2"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiDatabaseRefresh} className="mr-1" />
                                    Reclone
                                </MenuItem>
                                <MenuItem
                                    as={Button}
                                    disabled={!repoClonedAndHealthy(node)}
                                    onSelect={() => navigate(`/${node.name}/-/settings`)}
                                    className="p-2"
                                >
                                    <Icon aria-hidden={true} svgPath={mdiCog} className="mr-1" />
                                    Settings
                                </MenuItem>
                            </MenuList>
                        </Menu>
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

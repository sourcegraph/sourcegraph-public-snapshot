import React, { useCallback } from 'react'

import {
    mdiCloudOutline,
    mdiCheck,
    mdiGithub,
    mdiGitlab,
    mdiBitbucket,
    mdiSourceRepository,
    mdiChevronRight,
} from '@mdi/js'
import classNames from 'classnames'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { Badge, LoadingSpinner, Link, Icon, Checkbox, Tooltip } from '@sourcegraph/wildcard'

import { ExternalServiceKind } from '../../../graphql-operations'

import { RepositoryNodeContainer } from './components'

import styles from './RepositoryNode.module.scss'

interface RepositoryNodeProps {
    name: string
    mirrorInfo?: {
        cloneInProgress: boolean
        cloned: boolean
    }
    onClick?: () => void
    url: string
    serviceType: string
    isPrivate: boolean
    prefixComponent?: JSX.Element
}

interface StatusIconProps {
    mirrorInfo?: {
        cloneInProgress: boolean
        cloned: boolean
    }
}

const StatusIcon: React.FunctionComponent<React.PropsWithChildren<StatusIconProps>> = ({ mirrorInfo }) => {
    if (mirrorInfo === undefined) {
        return null
    }
    if (mirrorInfo.cloneInProgress) {
        return (
            <Tooltip content="Clone in progress.">
                <small className="mr-2 text-success">
                    <LoadingSpinner />
                </small>
            </Tooltip>
        )
    }
    if (!mirrorInfo.cloned) {
        return (
            <Tooltip content="Visit the repository to clone it. See its mirroring settings for diagnostics.">
                <small className="mr-2 text-muted">
                    <Icon aria-hidden={true} svgPath={mdiCloudOutline} />
                </small>
            </Tooltip>
        )
    }
    return (
        <small className="mr-2">
            <Icon className={styles.check} aria-label="Success" svgPath={mdiCheck} />
        </small>
    )
}

interface CodeHostIconProps {
    hostType: string
}

const CodeHostIcon: React.FunctionComponent<React.PropsWithChildren<CodeHostIconProps>> = ({ hostType }) => {
    switch (hostType) {
        case ExternalServiceKind.GITHUB:
            return (
                <small className="mr-2">
                    <Icon className={styles.github} aria-hidden={true} svgPath={mdiGithub} />
                </small>
            )
        case ExternalServiceKind.GITLAB:
            return (
                <small className="mr-2">
                    <Icon className={styles.gitlab} aria-hidden={true} svgPath={mdiGitlab} />
                </small>
            )
        case ExternalServiceKind.BITBUCKETCLOUD:
            return (
                <small className="mr-2">
                    <Icon aria-hidden={true} svgPath={mdiBitbucket} />
                </small>
            )
        default:
            return (
                <small className="mr-2">
                    <Icon aria-hidden={true} svgPath={mdiSourceRepository} />
                </small>
            )
    }
}

export const RepositoryNode: React.FunctionComponent<React.PropsWithChildren<RepositoryNodeProps>> = ({
    name,
    mirrorInfo,
    url,
    onClick,
    serviceType,
    isPrivate,
    prefixComponent,
}) => {
    const handleOnClick = useCallback(
        (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>): void => {
            if (onClick !== undefined) {
                event.preventDefault()
                onClick()
            }
        },
        [onClick]
    )

    return (
        <RepositoryNodeContainer as="tr">
            <td className="border-color">
                <Link
                    className={classNames('w-100 d-flex justify-content-between align-items-center', styles.link)}
                    to={url}
                    onClick={handleOnClick}
                >
                    <div className="d-flex align-items-center">
                        {prefixComponent && prefixComponent}
                        <StatusIcon mirrorInfo={mirrorInfo} />
                        <CodeHostIcon hostType={serviceType} />
                        <RepoLink className="text-muted" repoName={name} to={null} />
                    </div>
                    <div>
                        {isPrivate && (
                            <Badge variant="secondary" className="text-muted" as="div">
                                Private
                            </Badge>
                        )}
                        <Icon className="ml-2 text-primary" aria-hidden={true} svgPath={mdiChevronRight} />
                    </div>
                </Link>
            </td>
        </RepositoryNodeContainer>
    )
}

interface CheckboxRepositoryNodeProps {
    name: string
    mirrorInfo?: {
        cloneInProgress: boolean
        cloned: boolean
    }
    onClick: () => void
    checked: boolean
    serviceType: string
    isPrivate: boolean
}

export const CheckboxRepositoryNode: React.FunctionComponent<React.PropsWithChildren<CheckboxRepositoryNodeProps>> = ({
    name,
    mirrorInfo,
    onClick,
    checked,
    serviceType,
    isPrivate,
}) => {
    const handleOnClick = useCallback(
        (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>): void => {
            if (onClick !== undefined) {
                event.preventDefault()
                onClick()
            }
        },
        [onClick]
    )
    return (
        <tr className="cursor-pointer" key={name}>
            <RepositoryNodeContainer
                as="td"
                role="gridcell"
                className="p-2 w-100 d-flex justify-content-between"
                onClick={onClick}
            >
                <div className="d-flex">
                    <Checkbox
                        className="mr-3"
                        aria-label={`select ${name} repository`}
                        onChange={onClick}
                        checked={checked}
                        label={
                            <>
                                <StatusIcon mirrorInfo={mirrorInfo} />
                                <CodeHostIcon hostType={serviceType} />
                                <RepoLink
                                    className="text-muted"
                                    repoClassName="text-body"
                                    repoName={name}
                                    to={null}
                                    onClick={handleOnClick}
                                />
                            </>
                        }
                    />
                </div>
                <div>
                    {isPrivate && (
                        <Badge className="bg-color-2 text-muted" as="div">
                            Private
                        </Badge>
                    )}
                </div>
            </RepositoryNodeContainer>
        </tr>
    )
}

import React, { useCallback } from 'react'

import classNames from 'classnames'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import TickIcon from 'mdi-react/TickIcon'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { Badge, LoadingSpinner, Link, Icon, Checkbox } from '@sourcegraph/wildcard'

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
            <small data-tooltip="Clone in progress." className="mr-2 text-success">
                <LoadingSpinner />
            </small>
        )
    }
    if (!mirrorInfo.cloned) {
        return (
            <small
                className="mr-2 text-muted"
                data-tooltip="Visit the repository to clone it. See its mirroring settings for diagnostics."
                aria-label="Visit the repository to clone it. See its mirroring settings for diagnostics."
            >
                <Icon role="img" as={CloudOutlineIcon} aria-hidden={true} />
            </small>
        )
    }
    return (
        <small className="mr-2">
            <Icon role="img" className={styles.check} as={TickIcon} aria-label="Success" />
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
                    <Icon role="img" className={styles.github} as={GithubIcon} aria-hidden={true} />
                </small>
            )
        case ExternalServiceKind.GITLAB:
            return (
                <small className="mr-2">
                    <Icon role="img" className={styles.gitlab} as={GitlabIcon} aria-hidden={true} />
                </small>
            )
        case ExternalServiceKind.BITBUCKETCLOUD:
            return (
                <small className="mr-2">
                    <Icon role="img" as={BitbucketIcon} aria-hidden={true} />
                </small>
            )
        default:
            return (
                <small className="mr-2">
                    <Icon role="img" as={SourceRepositoryIcon} aria-hidden={true} />
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
                        <Icon role="img" className="ml-2 text-primary" as={ChevronRightIcon} aria-hidden={true} />
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

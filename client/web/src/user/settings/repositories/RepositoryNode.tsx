import classNames from 'classnames'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CloudOutlineIcon from 'mdi-react/CloudOutlineIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import TickIcon from 'mdi-react/TickIcon'
import React, { useCallback } from 'react'

import { RepoLink } from '@sourcegraph/shared/src/components/RepoLink'
import { Badge, LoadingSpinner } from '@sourcegraph/wildcard'

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

const StatusIcon: React.FunctionComponent<StatusIconProps> = ({ mirrorInfo }) => {
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
            >
                <CloudOutlineIcon className="icon-inline" />
            </small>
        )
    }
    return (
        <small className="mr-2">
            <TickIcon className={classNames('icon-inline', styles.check)} />
        </small>
    )
}

interface CodeHostIconProps {
    hostType: string
}

const CodeHostIcon: React.FunctionComponent<CodeHostIconProps> = ({ hostType }) => {
    switch (hostType) {
        case ExternalServiceKind.GITHUB:
            return (
                <small className="mr-2">
                    <GithubIcon className={classNames('icon-inline', styles.github)} />
                </small>
            )
        case ExternalServiceKind.GITLAB:
            return (
                <small className="mr-2">
                    <GitlabIcon className={classNames('icon-inline', styles.gitlab)} />
                </small>
            )
        case ExternalServiceKind.BITBUCKETCLOUD:
            return (
                <small className="mr-2">
                    <BitbucketIcon className="icon-inline" />
                </small>
            )
        default:
            return (
                <small className="mr-2">
                    <SourceRepositoryIcon className="icon-inline" />
                </small>
            )
    }
}

export const RepositoryNode: React.FunctionComponent<RepositoryNodeProps> = ({
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
                <a
                    className={classNames('w-100 d-flex justify-content-between align-items-center', styles.link)}
                    href={url}
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
                        <ChevronRightIcon className="icon-inline ml-2 text-primary" />
                    </div>
                </a>
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

export const CheckboxRepositoryNode: React.FunctionComponent<CheckboxRepositoryNodeProps> = ({
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
                <div className="d-flex align-items-center">
                    <input
                        className="mr-3"
                        type="checkbox"
                        aria-label={`select ${name} repository`}
                        onChange={onClick}
                        checked={checked}
                    />
                    <StatusIcon mirrorInfo={mirrorInfo} />
                    <CodeHostIcon hostType={serviceType} />
                    <RepoLink
                        className="text-muted"
                        repoClassName="text-body"
                        repoName={name}
                        to={null}
                        onClick={handleOnClick}
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

import React, { useState, useMemo } from 'react'

import { mdiInformation, mdiAlert, mdiSync, mdiCheckboxMarkedCircle } from '@mdi/js'
import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery } from '@sourcegraph/http-client'
import {
    CloudAlertIconRefresh,
    CloudSyncIconRefresh,
    CloudCheckIconRefresh,
} from '@sourcegraph/shared/src/components/icons'
import {
    Button,
    Link,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Icon,
    H4,
    Text,
    Tooltip,
} from '@sourcegraph/wildcard'

import { StatusMessagesResult } from '../graphql-operations'

import { STATUS_MESSAGES } from './StatusMessagesNavItemQueries'

import styles from './StatusMessagesNavItem.module.scss'

type EntryType = 'progress' | 'warning' | 'success' | 'error'

interface StatusMessageEntryProps {
    title: string
    message: string
    linkTo: string
    linkText: string
    entryType: EntryType
    linkOnClick: (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => void
    messageHint?: string
}

function entryIcon(entryType: EntryType): JSX.Element {
    const sharedProps = { height: 14, width: 14, inline: false }

    switch (entryType) {
        case 'error':
            return (
                <Icon
                    {...sharedProps}
                    className={classNames('text-danger', styles.icon)}
                    svgPath={mdiInformation}
                    aria-label="Error"
                />
            )
        case 'warning':
            return (
                <Icon
                    {...sharedProps}
                    className={classNames('text-warning', styles.icon)}
                    svgPath={mdiAlert}
                    aria-label="Warning"
                />
            )
        case 'success':
            return (
                <Icon
                    {...sharedProps}
                    className={classNames('text-success', styles.icon)}
                    svgPath={mdiCheckboxMarkedCircle}
                    aria-label="Success"
                />
            )
        case 'progress':
            return (
                <Icon
                    {...sharedProps}
                    className={classNames('text-primary', styles.icon)}
                    svgPath={mdiSync}
                    aria-label="In progress"
                />
            )
    }
}

const getMessageColor = (entryType: EntryType): string => {
    switch (entryType) {
        case 'error':
            return styles.messageError
        case 'warning':
            return styles.messageWarning
        default:
            return ''
    }
}

const getBorderClassname = (entryType: EntryType): string => {
    switch (entryType) {
        case 'error':
            return styles.entryBorderError
        case 'warning':
            return styles.entryBorderWarning
        case 'success':
            return styles.entryBorderSuccess
        case 'progress':
            return styles.entryBorderProgress
        default:
            return ''
    }
}

const StatusMessagesNavItemEntry: React.FunctionComponent<React.PropsWithChildren<StatusMessageEntryProps>> = props => (
    <div key={props.message} className={styles.entry}>
        <H4 className="d-flex align-items-center mb-0">
            {entryIcon(props.entryType)}
            {props.title ? props.title : 'Your repositories'}
        </H4>
        <div
            className={classNames(
                'status-messages-nav-item__entry-card',
                styles.cardActive,
                getBorderClassname(props.entryType)
            )}
        >
            <Text className={classNames(styles.message, getMessageColor(props.entryType))}>{props.message}</Text>
            {props.messageHint && (
                <>
                    <small className="text-muted d-inline-block mb-1">{props.messageHint}</small>
                    <br />
                </>
            )}
            <Link className="text-primary" to={props.linkTo} onClick={props.linkOnClick}>
                {props.linkText}
            </Link>
        </div>
    </div>
)

const STATUS_MESSAGES_POLL_INTERVAL = 10000

interface Props {
    disablePolling?: boolean
}
/**
 * Displays a status icon in the navbar reflecting the completion of backend
 * tasks such as repository cloning, and exposes a dropdown menu containing
 * more information on these tasks.
 */
export const StatusMessagesNavItem: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = (): void => setIsOpen(old => !old)

    const { data, error } = useQuery<StatusMessagesResult>(STATUS_MESSAGES, {
        fetchPolicy: 'no-cache',
        pollInterval:
            props.disablePolling !== undefined && !props.disablePolling ? STATUS_MESSAGES_POLL_INTERVAL : undefined,
    })

    const icon: JSX.Element | null = useMemo(() => {
        if (!data) {
            return null
        }

        let codeHostMessage
        let icon
        if (
            data.statusMessages?.some(
                ({ __typename: type }) => type === 'ExternalServiceSyncError' || type === 'SyncError'
            )
        ) {
            codeHostMessage = 'Syncing repositories failed!'
            icon = CloudAlertIconRefresh
        } else if (data.statusMessages?.some(({ __typename: type }) => type === 'CloningProgress')) {
            codeHostMessage = 'Cloning repositories...'
            icon = CloudSyncIconRefresh
        } else {
            codeHostMessage = 'Repositories up to date'
            icon = CloudCheckIconRefresh
        }

        return (
            <Tooltip content={isOpen ? undefined : codeHostMessage}>
                <Icon as={icon} size="md" {...(isOpen ? { 'aria-hidden': true } : { 'aria-label': codeHostMessage })} />
            </Tooltip>
        )
    }, [data, isOpen])

    const messages: JSX.Element | null = useMemo(() => {
        if (!data) {
            return null
        }

        // no status messages
        if (data.statusMessages.length === 0) {
            return (
                <StatusMessagesNavItemEntry
                    key="up-to-date"
                    title="Repositories up to date"
                    message="Repositories synced from code host and available for search."
                    linkTo="/site-admin/repositories"
                    linkText="View repositories"
                    linkOnClick={toggleIsOpen}
                    entryType="success"
                />
            )
        }

        return (
            <>
                {data.statusMessages.map(status => {
                    if (status.__typename === 'CloningProgress') {
                        return (
                            <StatusMessagesNavItemEntry
                                key={status.message}
                                message={status.message}
                                title="Cloning repositories"
                                messageHint="Not all repositories available for search yet."
                                linkTo="/site-admin/repositories"
                                linkText="View repositories"
                                linkOnClick={toggleIsOpen}
                                entryType="progress"
                            />
                        )
                    }
                    if (status.__typename === 'ExternalServiceSyncError') {
                        return (
                            <StatusMessagesNavItemEntry
                                key={status.externalService.id}
                                title="Code host connection"
                                message={`Failed to connect to "${status.externalService.displayName}".`}
                                messageHint="Repositories synced to Sourcegraph may not be up to date."
                                linkTo={`/site-admin/external-services/${status.externalService.id}`}
                                linkText="View code host configuration"
                                linkOnClick={toggleIsOpen}
                                entryType="error"
                            />
                        )
                    }
                    if (status.__typename === 'SyncError') {
                        return (
                            <StatusMessagesNavItemEntry
                                key={status.message}
                                message={status.message}
                                title="Syncing repositories from code hosts"
                                messageHint="Repository contents may not be up to date."
                                linkTo="/site-admin/repositories?status=failed-fetch"
                                linkText="View affected repositories"
                                linkOnClick={toggleIsOpen}
                                entryType="warning"
                            />
                        )
                    }
                    return null
                })}
            </>
        )
    }, [data])

    return (
        <Popover isOpen={isOpen} onOpenChange={event => setIsOpen(event.isOpen)}>
            <PopoverTrigger
                className="nav-link py-0 px-0 percy-hide chromatic-ignore"
                as={Button}
                variant="link"
                aria-label={isOpen ? 'Hide status messages' : 'Show status messages'}
            >
                {error && (
                    <Tooltip content="Sorry, we couldn’t fetch notifications!">
                        <Icon
                            aria-label="Sorry, we couldn’t fetch notifications!"
                            as={CloudAlertIconRefresh}
                            size="md"
                        />
                    </Tooltip>
                )}
                {icon}
            </PopoverTrigger>

            <PopoverContent position={Position.bottomEnd} className={classNames('p-0', styles.dropdownMenu)}>
                <div className={styles.dropdownMenuContent}>
                    <small className={classNames('d-inline-block text-muted', styles.sync)}>Status</small>
                    {error && (
                        <ErrorAlert className={styles.entry} prefix="Failed to load status messages" error={error} />
                    )}
                    {messages}
                </div>
            </PopoverContent>
        </Popover>
    )
}

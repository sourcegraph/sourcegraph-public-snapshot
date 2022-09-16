import React from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { drop } from 'lodash'
import { useLocation } from 'react-router'

import { Alert, Button, Icon } from '@sourcegraph/wildcard'

import styles from './InviteMemberModal.module.scss'

export const getInvitationExpiryDateString = (expiring: string): string => {
    const expiryDate = new Date(expiring)
    const now = new Date().getTime()
    const diff = expiryDate.getTime() - now
    const numberOfDays = diff / (1000 * 3600 * 24)
    if (numberOfDays < 1) {
        return 'today'
    }

    const numberDaysInt = Math.round(numberOfDays)

    if (numberDaysInt === 1) {
        return 'tomorrow'
    }

    return `in ${numberDaysInt} days`
}

export const getInvitationCreationDateString = (creation: string): string => {
    const creationDate = new Date(creation)
    const now = new Date().getTime()
    const diff = now - creationDate.getTime()
    const numberOfDays = diff / (1000 * 3600 * 24)
    const numberDaysInt = Math.round(numberOfDays)
    if (numberDaysInt < 1) {
        return 'today'
    }

    if (numberDaysInt === 1) {
        return 'yesterday'
    }

    return `${numberDaysInt} days ago`
}

export const getLocaleFormattedDateFromString = (jsonDate: string): string => new Date(jsonDate).toLocaleDateString()

interface MembersNotificationProps {
    message: string
    onDismiss: () => void
    className?: string
}

export const OrgMemberNotification: React.FunctionComponent<React.PropsWithChildren<MembersNotificationProps>> = ({
    className,
    message,
    onDismiss,
}) => (
    <Alert variant="success" className={classNames(styles.invitedNotification, className)}>
        <div className={styles.message}>{message}</div>
        <Button title="Dismiss" onClick={onDismiss}>
            <Icon aria-hidden={true} svgPath={mdiClose} />
        </Button>
    </Alert>
)

export function getPaginatedItems<T>(
    currentPage: number,
    items?: T[],
    pageSize = 20
): { totalPages: number; results: T[] } {
    if (!items || items.length === 0) {
        return {
            totalPages: 0,
            results: [],
        }
    }
    const page = currentPage || 1
    const offset = (page - 1) * pageSize
    const pagedItems = drop(items, offset).slice(0, pageSize)
    return {
        totalPages: Math.ceil(items.length / pageSize),
        results: pagedItems,
    }
}

export function useQueryStringParameters(): URLSearchParams {
    const { search } = useLocation()

    return React.useMemo(() => new URLSearchParams(search), [search])
}

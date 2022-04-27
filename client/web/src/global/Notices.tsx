import React, { useMemo } from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { Notice, Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Alert, AlertProps } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { DismissibleAlert } from '../components/DismissibleAlert'

import styles from './Notices.module.scss'

const getAlertVariant = (location: Notice['location']): AlertProps['variant'] =>
    location === 'top' ? 'info' : undefined

interface NoticeAlertProps {
    notice: Notice
    className?: string
    testId?: string
}

const NoticeAlert: React.FunctionComponent<React.PropsWithChildren<NoticeAlertProps>> = ({
    notice,
    className = '',
    testId,
}) => {
    const content = <Markdown dangerousInnerHTML={renderMarkdown(notice.message)} />

    const sharedProps = {
        'data-testid': testId,
        variant: getAlertVariant(notice.location),
        className: classNames(notice.location !== 'top' && 'bg transparent border p-2', className),
    }

    return notice.dismissible ? (
        <DismissibleAlert {...sharedProps} partialStorageKey={`notice.${notice.message}`}>
            {content}
        </DismissibleAlert>
    ) : (
        <Alert {...sharedProps}>{content}</Alert>
    )
}

interface Props extends SettingsCascadeProps {
    className?: string

    /** Apply this class name to each notice (alongside .alert). */
    alertClassName?: string

    /** Display notices for this location. */
    location: Notice['location']

    authenticatedUser: AuthenticatedUser | null
}

// FIXME: change to get from featureFlags
const experimentEnabled = true

/**
 * Displays notices from settings for a specific location.
 */
export const Notices: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className = '',
    alertClassName,
    settingsCascade,
    location,
    authenticatedUser,
}) => {
    const notices = useMemo(() => {
        const verifyEmailNotices = (experimentEnabled && authenticatedUser ? authenticatedUser.emails : [])
            .filter(({ verified }) => !verified)
            .map(
                ({ email }): Notice => ({
                    message: `Your email ${email as string} is not verified. <a href="#">Send verification email.</a>`,
                    location: 'top',
                    dismissible: false,
                })
            )

        if (
            !isSettingsValid<Settings>(settingsCascade) ||
            !settingsCascade.final.notices ||
            !Array.isArray(settingsCascade.final.notices)
        ) {
            return verifyEmailNotices
        }

        return [...verifyEmailNotices, ...settingsCascade.final.notices.filter(notice => notice.location === location)]
    }, [authenticatedUser, location, settingsCascade])

    if (notices.length === 0) {
        return null
    }

    return (
        <div className={classNames(styles.notices, className)}>
            {notices.map((notice, index) => (
                <NoticeAlert key={index} testId="notice-alert" className={alertClassName} notice={notice} />
            ))}
        </div>
    )
}

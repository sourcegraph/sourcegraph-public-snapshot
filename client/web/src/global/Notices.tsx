import React, { useMemo } from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/common'
import { Notice } from '@sourcegraph/shared/src/schema/settings.schema'
import { useSettings } from '@sourcegraph/shared/src/settings/settings'
import { Alert, AlertProps, Markdown } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../components/DismissibleAlert'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

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

interface Props {
    className?: string

    /** Apply this class name to each notice (alongside .alert). */
    alertClassName?: string

    /** Display notices for this location. */
    location: Notice['location']
}

/**
 * Displays notices from settings for a specific location.
 */
export const Notices: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className = '',
    alertClassName,
    location,
}) => {
    const settings = useSettings()

    if (!settings?.notices || !Array.isArray(settings?.notices)) {
        return null
    }

    const notices = settings.notices.filter(notice => notice.location === location)
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

interface VerifyEmailNoticesProps {
    className?: string
    /** Apply this class name to each notice (alongside .alert). */
    alertClassName?: string
    emails: string[]
    settingsURL: string
}

/**
 * Displays notices from settings for a specific location.
 */
export const VerifyEmailNotices: React.FunctionComponent<VerifyEmailNoticesProps> = ({
    className,
    alertClassName,
    emails,
    settingsURL,
}) => {
    const [isEmailVerificationAlertEnabled, status] = useFeatureFlag('ab-email-verification-alert')

    const notices: Notice[] = useMemo(() => {
        if (status !== 'loaded' || !isEmailVerificationAlertEnabled) {
            return []
        }
        return emails.map(
            (email): Notice => ({
                message: `Please, <a href="${settingsURL}/emails">verify your email</a> <strong>${email
                    .split('@')
                    .join('\\@')}</strong>`,
                location: 'top',
                dismissible: false,
            })
        )
    }, [emails, isEmailVerificationAlertEnabled, settingsURL, status])

    if (notices.length === 0) {
        return null
    }

    return (
        <div className={classNames(styles.notices, className)}>
            {notices.map(notice => (
                <NoticeAlert key={notice.message} testId="notice-alert" className={alertClassName} notice={notice} />
            ))}
        </div>
    )
}

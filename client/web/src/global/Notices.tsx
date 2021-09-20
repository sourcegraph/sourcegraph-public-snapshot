import classNames from 'classnames'
import * as React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { DismissibleAlert } from '../components/DismissibleAlert'
import { Notice, Settings } from '../schema/settings.schema'

import styles from './Notices.module.scss'

const NoticeAlert: React.FunctionComponent<{ notice: Notice; className?: string; testId?: string }> = ({
    notice,
    className = '',
    testId,
}) => {
    const content = <Markdown dangerousInnerHTML={renderMarkdown(notice.message)} />
    const baseClassName = notice.location === 'top' ? 'alert-info' : 'bg-transparent border'

    return notice.dismissible ? (
        <DismissibleAlert
            data-testid={testId}
            className={classNames(baseClassName, className)}
            partialStorageKey={`notice.${notice.message}`}
        >
            {content}
        </DismissibleAlert>
    ) : (
        <div data-testid={testId} className={classNames('alert', baseClassName, className)}>
            {content}
        </div>
    )
}

interface Props extends SettingsCascadeProps {
    className?: string

    /** Apply this class name to each notice (alongside .alert). */
    alertClassName?: string

    /** Display notices for this location. */
    location: Notice['location']
}

/**
 * Displays notices from settings for a specific location.
 */
export const Notices: React.FunctionComponent<Props> = ({
    className = '',
    alertClassName,
    settingsCascade,
    location,
}) => {
    if (
        !isSettingsValid<Settings>(settingsCascade) ||
        !settingsCascade.final.notices ||
        !Array.isArray(settingsCascade.final.notices)
    ) {
        return null
    }

    const notices = settingsCascade.final.notices.filter(notice => notice.location === location)
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

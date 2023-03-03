import React from 'react'

import classNames from 'classnames'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TemporarySettingsSchema } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { Badge, Button, Text } from '@sourcegraph/wildcard'

import styles from './LimitedAccessBanner.module.scss'

interface LimitedAccessBannerProps extends React.HTMLAttributes<HTMLDivElement> {
    badgeText?: string
    dismissableTemporarySettingsKey: keyof TemporarySettingsSchema
}

export const LimitedAccessBanner: React.FunctionComponent<
    React.PropsWithChildren<LimitedAccessBannerProps>
> = props => {
    const badgeText = props.badgeText ?? 'Limited access'
    const [dismissed, setDismissed] = useTemporarySetting(props.dismissableTemporarySettingsKey)

    if (dismissed) {
        return (
            <Button variant="merged" onClick={() => setDismissed(false)}>
                Undismiss
            </Button>
        )
    }

    return (
        <div {...props} className={classNames(styles.banner, props.className, 'my-4')}>
            <div className={styles.content}>
                <div>
                    <Badge className={classNames('mb-2', styles.badge)}>{badgeText}</Badge>
                    <Text className="m-0">{props.children}</Text>
                </div>
                <Button className={styles.dismiss} variant="link" onClick={() => setDismissed(true)}>
                    Dismiss
                </Button>
            </div>
        </div>
    )
}

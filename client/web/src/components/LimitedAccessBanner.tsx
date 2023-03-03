import React from 'react'

import classNames from 'classnames'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TemporarySettingsSchema } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { Badge, Button, Text } from '@sourcegraph/wildcard'

import styles from './LimitedAccessBanner.module.scss'

interface LimitedAccessBannerProps {
    badgeText?: string
    dismissableTemporarySettingsKey: keyof TemporarySettingsSchema
}

export const LimitedAccessBanner: React.FunctionComponent<
    React.PropsWithChildren<LimitedAccessBannerProps>
> = props => {
    const badgeText = props.badgeText ?? 'Limited access'
    const [dismissed, setDismissed] = useTemporarySetting(props.dismissableTemporarySettingsKey)

    if (dismissed) {
        return null
    }

    return (
        <div className={classNames('my-4 p-1', styles.banner)}>
            <div className={classNames('py-2 px-3', styles.content)}>
                <div>
                    <Badge className={classNames('mb-2', styles.badge)}>{badgeText}</Badge>
                    <Text className="m-0">{props.children}</Text>
                </div>
                <Button className="align-self-start" variant="link" onClick={() => setDismissed(true)}>
                    Dismiss
                </Button>
            </div>
        </div>
    )
}

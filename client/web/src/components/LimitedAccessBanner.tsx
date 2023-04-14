import React from 'react'

import classNames from 'classnames'

import { Badge, Button, Text, useLocalStorage } from '@sourcegraph/wildcard'

import styles from './LimitedAccessBanner.module.scss'

interface LimitedAccessBannerProps {
    badgeText?: string
    storageKey: string
    className?: string
}

export const LimitedAccessBanner: React.FunctionComponent<
    React.PropsWithChildren<LimitedAccessBannerProps>
> = props => {
    const badgeText = props.badgeText ?? 'Limited access'
    const [dismissed, setDismissed] = useLocalStorage<boolean>(props.storageKey, false)

    if (dismissed) {
        return null
    }

    return (
        <div className={classNames('p-1', styles.banner, props.className)}>
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

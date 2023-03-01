import React from 'react'

import classNames from 'classnames'

import { Badge, Text } from '@sourcegraph/wildcard'

import styles from './LimitedAccessBanner.module.scss'

interface LimitedAccessBannerProps extends React.HTMLAttributes<HTMLDivElement> {}

export const LimitedAccessBanner: React.FunctionComponent<
    React.PropsWithChildren<LimitedAccessBannerProps>
> = props => (
    <div {...props} className={classNames(styles.banner, props.className, 'my-4')}>
        <div className={styles.content}>
            <Badge className={classNames('mb-2', styles.badge)}>LIMITED ACCESS</Badge>
            <Text className="m-0">{props.children}</Text>
        </div>
    </div>
)

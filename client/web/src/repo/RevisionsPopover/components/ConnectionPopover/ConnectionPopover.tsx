import React, { type HTMLAttributes } from 'react'

import classNames from 'classnames'

import { Tabs, type TabsProps } from '@sourcegraph/wildcard'

import styles from './ConnectionPopover.module.scss'

type ConnectionPopoverProps = HTMLAttributes<HTMLDivElement>

export const ConnectionPopover: React.FunctionComponent<React.PropsWithChildren<ConnectionPopoverProps>> = ({
    className,
    children,
    ...rest
}) => (
    <div className={classNames(styles.connectionPopover, className)} {...rest}>
        {children}
    </div>
)

type ConnectionPopoverTabsProps = TabsProps & {
    className?: string
}

export const ConnectionPopoverTabs: React.FunctionComponent<React.PropsWithChildren<ConnectionPopoverTabsProps>> = ({
    children,
    className,
    ...rest
}) => (
    <div className={classNames(styles.connectionPopover, className)}>
        <Tabs {...rest} className={styles.tabs}>
            {children}
        </Tabs>
    </div>
)

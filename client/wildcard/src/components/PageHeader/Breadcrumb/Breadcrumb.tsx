import React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import styles from './Breadcrumb.module.scss'

export type BreadcrumbIcon = React.ComponentType<{ className?: string; role?: React.AriaRole }>
export type BreadcrumbText = React.ReactNode

type BreadcrumbProps = React.HTMLAttributes<HTMLSpanElement> & {
    /** Use a valid path to render this Breadcrumb as a Link */
    to?: string
    icon?: BreadcrumbIcon
    children?: React.ReactNode
} & (
        | {
              icon: BreadcrumbIcon
          }
        | {
              children: BreadcrumbText
          }
    )

export const Breadcrumb: React.FunctionComponent<BreadcrumbProps> = ({
    to,
    icon: Icon,
    className,
    children,
    'aria-label': ariaLabel,
    ...rest
}) => {
    const iconHidden = !!children || !ariaLabel

    return (
        <span className={classNames(styles.wrapper, className)} {...rest}>
            <LinkOrSpan className={styles.path} to={to} aria-label={children ? ariaLabel : undefined}>
                {Icon && (
                    <Icon
                        role="img"
                        className={styles.icon}
                        aria-hidden={iconHidden}
                        aria-label={iconHidden ? undefined : ariaLabel}
                    />
                )}
                {children && <span className={styles.text}>{children}</span>}
            </LinkOrSpan>
        </span>
    )
}

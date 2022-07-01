import React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'

import { Icon } from '../../Icon'

import styles from './Breadcrumb.module.scss'

export type BreadcrumbIcon = React.ComponentType<{ className?: string }> | string
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
    icon,
    className,
    children,
    'aria-label': ariaLabel,
    ...rest
}) => {
    const iconHidden = !!children || !ariaLabel

    return (
        <span className={classNames(styles.wrapper, className)} {...rest}>
            <LinkOrSpan className={styles.path} to={to} aria-label={children ? ariaLabel : undefined}>
                {icon && (
                    <Icon
                        className={styles.icon}
                        svgPath={typeof icon === 'string' ? icon : undefined}
                        as={typeof icon !== 'string' ? icon : undefined}
                        {...(iconHidden ? { 'aria-hidden': true } : { 'aria-label': ariaLabel })}
                    />
                )}
                {children && <span className={styles.text}>{children}</span>}
            </LinkOrSpan>
        </span>
    )
}

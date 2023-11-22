import * as React from 'react'

import classNames from 'classnames'

import { ButtonLink, type ButtonLinkProps, Tooltip } from '@sourcegraph/wildcard'

import styles from './SimpleActionItem.module.scss'

interface SimpleActionItemProps extends Omit<ButtonLinkProps, 'href'> {
    isActive?: boolean
    tooltip: string
}

export const SimpleActionItem: React.FunctionComponent<SimpleActionItemProps> = props => {
    const { isActive, tooltip, children, className, ...buttonLinkProps } = props

    return (
        <div className={styles.margin}>
            <Tooltip content={props.tooltip} placement="left">
                <ButtonLink
                    aria-label={props.tooltip}
                    className={classNames(
                        styles.simpleActionItem,
                        isActive && styles.simpleActionItemActive,
                        className
                    )}
                    {...buttonLinkProps}
                >
                    {children}
                </ButtonLink>
            </Tooltip>
        </div>
    )
}

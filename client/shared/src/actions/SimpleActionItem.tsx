import * as React from 'react'

import classNames from 'classnames'

import { ButtonLink, ButtonLinkProps, Tooltip } from '@sourcegraph/wildcard'

import styles from './SimpleActionItem.module.scss'

interface SimpleActionItemProps extends Omit<ButtonLinkProps, 'href'> {
    isActive?: boolean
    tooltip: string
}

export const SimpleActionItem: React.FunctionComponent<SimpleActionItemProps> = props => {
    const { isActive, tooltip, children, className, ...buttonLinkProps } = props

    return (
        <Tooltip content={props.tooltip}>
            <span>
                {/**
                 * This <ButtonLink> must be wrapped with an additional span, since the tooltip currently has an issue that will
                 * break its onClick handler, and it will no longer prevent the default page reload (with no href).
                 */}
                <ButtonLink
                    className={classNames(
                        styles.simpleActionItem,
                        isActive && styles.simpleActionItemActive,
                        className
                    )}
                    aria-label={props.tooltip}
                    {...buttonLinkProps}
                >
                    {children}
                </ButtonLink>
            </span>
        </Tooltip>
    )
}

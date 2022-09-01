import * as React from 'react'

import classNames from 'classnames'

import { Button, Tooltip } from '@sourcegraph/wildcard'

import styles from './SimpleActionItem.module.scss'

export interface SimpleActionItemProps {
    className: string
    iconURL: string
    tooltip: string
    onClick: () => void
}

export const SimpleActionItem: React.FunctionComponent<SimpleActionItemProps> = props => {
    const { className, iconURL, tooltip, onClick, ...otherProps } = props
    return (
        <Tooltip content={tooltip}>
            <Button
                className={classNames(className, styles.simpleActionItem)}
                onClick={onClick}
                aria-label={tooltip}
                {...otherProps}
            >
                <img src={iconURL} alt={tooltip || ''} />
            </Button>
        </Tooltip>
    )
}

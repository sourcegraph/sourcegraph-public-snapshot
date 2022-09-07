import * as React from 'react'

import classNames from 'classnames'

import { Button, Tooltip } from '@sourcegraph/wildcard'

import styles from './SimpleActionItem.module.scss'

interface Props {
    isActive?: boolean
    tooltip: string
    onClick: (event: React.MouseEvent<HTMLElement>) => void
}

type SimpleActionItemProps = Props & ({ icon: React.ReactElement } | { iconURL: string })

export const SimpleActionItem: React.FunctionComponent<SimpleActionItemProps> = props => {
    const icon =
        'icon' in props ? props.icon : 'iconURL' in props ? <img src={props.iconURL} alt={props.tooltip || ''} /> : null

    if (!icon) {
        return null
    }

    return (
        <Tooltip content={props.tooltip}>
            <Button
                className={classNames(styles.simpleActionItem, props.isActive && styles.simpleActionItemActive)}
                onClick={props.onClick}
                aria-label={props.tooltip}
            >
                {icon}
            </Button>
        </Tooltip>
    )
}

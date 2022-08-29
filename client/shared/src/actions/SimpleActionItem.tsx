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

export const SimpleActionItem: React.FunctionComponent<SimpleActionItemProps> = props => (
        <Tooltip content={props.tooltip}>
            <Button className={classNames(props.className, styles.simpleActionItem)} onClick={props.onClick} aria-label={props.tooltip}>
                <img src={props.iconURL} alt={props.tooltip || ''} />
            </Button>
        </Tooltip>
    )

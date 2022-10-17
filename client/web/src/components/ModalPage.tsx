import React from 'react'

import classNames from 'classnames'

import { CardBody, Card } from '@sourcegraph/wildcard'

import styles from './ModalPage.module.scss'

interface Props {
    icon?: React.ReactNode

    className?: string
    children?: React.ReactNode
}

/**
 * A page that displays a modal prompt in the middle of the screen.
 */
export const ModalPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    icon,
    className = '',
    children,
}) => (
    <div className={classNames(styles.modalPage, className)}>
        <Card>
            <CardBody>
                {icon && <div className={styles.icon}>{icon}</div>}
                {children}
            </CardBody>
        </Card>
    </div>
)

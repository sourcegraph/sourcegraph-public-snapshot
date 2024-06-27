import React from 'react'

import classNames from 'classnames'

import { AlertBadge } from './AlertBadge'
import { CodyProBadge } from './CodyProBadge'

import styles from './CodyAlert.module.scss'

interface CodyAlertProps extends React.HTMLAttributes<HTMLDivElement> {
    variant: 'purple' | 'green' | 'error'
    className?: string
    displayCard?: 'CodyPro' | 'Alert' | ''
    children: React.ReactNode
}

export const CodyAlert: React.FunctionComponent<CodyAlertProps> = ({
    variant,
    className,
    title,
    children,
    displayCard = '',
    ...props
}) => {
    const alertClassName = classNames(
        styles.alert,
        {
            [styles.purple]: variant === 'purple',
            [styles.green]: variant === 'green',
            [styles.error]: variant === 'error',
        },
        className
    )

    return (
        // eslint-disable-next-line no-restricted-syntax
        <div className={alertClassName} {...props}>
            {displayCard !== '' && (
                <div className={styles.card}>
                    <div className={styles.cardClip}>
                        {displayCard === 'CodyPro' ? <CodyProBadge /> : <AlertBadge />}
                    </div>
                </div>
            )}
            <div className={styles.content}>{children}</div>
        </div>
    )
}

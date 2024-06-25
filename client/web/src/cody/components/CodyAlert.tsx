import React from 'react'

import classNames from 'classnames'

import styles from './CodyAlert.module.scss'

interface CodyAlertProps extends React.HTMLAttributes<HTMLDivElement> {
    variant: 'purple' | 'greenSuccess' | 'purpleSuccess' | 'purpleCodyPro' | 'greenCodyPro' | 'error'
    className?: string
    children: React.ReactNode
}

export const CodyAlert: React.FunctionComponent<CodyAlertProps> = ({ variant, className, children, ...props }) => {
    const alertClassName = classNames(
        'mb-4',
        styles.alert,
        {
            [styles.purpleNoIcon]: variant === 'purple',
            [styles.greenSuccess]: variant === 'greenSuccess',
            [styles.purpleSuccess]: variant === 'purpleSuccess',
            [styles.purpleCodyPro]: variant === 'purpleCodyPro',
            [styles.greenCodyPro]: variant === 'greenCodyPro',
            [styles.error]: variant === 'error',
        },
        className
    )

    return (
        // eslint-disable-next-line no-restricted-syntax
        <div className={alertClassName} {...props}>
            {children}
        </div>
    )
}

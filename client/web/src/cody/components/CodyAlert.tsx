import React from 'react'

import classNames from 'classnames'

import styles from './CodyAlert.module.scss'

interface CodyAlertProps extends React.HTMLAttributes<HTMLDivElement> {
    variant: 'greenSuccess' | 'purpleSuccess' | 'error'
    className?: string
    children: React.ReactNode
}

export const CodyAlert: React.FunctionComponent<CodyAlertProps> = ({ variant, className, children, ...props }) => {
    const alertClassName = classNames(
        'mb-4',
        styles.alert,
        {
            [styles.greenSuccessAlert]: variant === 'greenSuccess',
            [styles.purpleSuccessAlert]: variant === 'purpleSuccess',
            [styles.errorAlert]: variant === 'error',
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

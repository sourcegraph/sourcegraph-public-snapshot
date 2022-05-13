import React from 'react'

import classNames from 'classnames'

import { useWildcardTheme } from '../../hooks'
import { ForwardReferenceComponent } from '../../types'

import { ALERT_VARIANTS } from './constants'
import { getAlertStyle } from './utils'

import styles from './Alert.module.scss'

type AlertVariant = typeof ALERT_VARIANTS[number]

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: AlertVariant
}

const userShouldBeNotified = (variant?: AlertVariant): boolean => variant === 'danger' || variant === 'warning'

export const Alert = React.forwardRef(
    ({ children, as: Component = 'div', variant, className, role, ...attributes }, reference) => {
        const { isBranded } = useWildcardTheme()
        const brandedClassName = isBranded && classNames(styles.alert, variant && getAlertStyle({ variant }))
        const alertRole = role || userShouldBeNotified(variant) ? 'alert' : undefined

        return (
            <Component
                ref={reference}
                className={classNames(brandedClassName, className)}
                role={alertRole}
                {...attributes}
            >
                {children}
            </Component>
        )
    }
) as ForwardReferenceComponent<'div', AlertProps>

import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

import styles from './Alert.module.scss'
import { ALERT_VARIANTS } from './constants'
import { getAlertStyle } from './utils'

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: typeof ALERT_VARIANTS[number]
    as?: React.ElementType

    /**
     * If the Alert should use branded styles. Defaults to true.
     */
    branded?: boolean
}

export const Alert = React.forwardRef(
    ({ children, as: Component = 'div', variant, className, branded = true, ...attributes }, reference) => {
        const brandedClassName = branded && classNames(styles.alert, variant && getAlertStyle({ variant }))

        return (
            <Component ref={reference} className={classNames(brandedClassName, className)} {...attributes}>
                {children}
            </Component>
        )
    }
) as ForwardReferenceComponent<'div', AlertProps>

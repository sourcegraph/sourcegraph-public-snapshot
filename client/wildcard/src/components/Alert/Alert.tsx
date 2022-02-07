import classNames from 'classnames'
import React from 'react'

import { useWildcardTheme } from '../../hooks'
import { ForwardReferenceComponent } from '../../types'

import styles from './Alert.module.scss'
import { ALERT_VARIANTS } from './constants'
import { getAlertStyle } from './utils'

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: typeof ALERT_VARIANTS[number]
}

export const Alert = React.forwardRef(
    ({ children, as: Component = 'div', variant, className, ...attributes }, reference) => {
        const { isBranded } = useWildcardTheme()
        const brandedClassName = isBranded && classNames(styles.alert, variant && getAlertStyle({ variant }))

        return (
            <Component ref={reference} className={classNames(brandedClassName, className)} {...attributes}>
                {children}
            </Component>
        )
    }
) as ForwardReferenceComponent<'div', AlertProps>

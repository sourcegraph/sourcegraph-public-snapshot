import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../..'

import { ICON_SIZES } from './constants'

import styles from './IconStyle.module.scss'

export interface IconStyleProps {
    className?: string
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
}

// eslint-disable-next-line react/display-name
export const IconStyle = React.forwardRef(
    ({ children, className, size, as: Component = 'div', ...attributes }, reference) => (
        <Component
            className={classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className)}
            ref={reference}
            {...attributes}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', IconStyleProps>

IconStyle.displayName = 'IconStyle'

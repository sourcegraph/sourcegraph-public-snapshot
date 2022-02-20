import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../..'

import { ICON_SIZES } from './constants'
import styles from './Icon.module.scss'

interface IconProps {
    className?: string
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
    /**
     * Used to change the element that is rendered.
     * Always be mindful of potentially accessibility pitfalls when using this!
     */
    as?: React.ElementType
}

export const Icon = React.forwardRef(
    ({ children, className, size, as: Component = 'div', ...attributes }, reference) => (
        <Component
            className={classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className)}
            ref={reference}
            {...attributes}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', IconProps>

import React, { ComponentType, ElementType, PropsWithChildren } from 'react'

import classNames from 'classnames'
import { MdiReactIconProps } from 'mdi-react'

import { ForwardReferenceComponent } from '../..'

import { ICON_SIZES } from './constants'

import styles from './Icon.module.scss'

export interface IconProps extends Omit<MdiReactIconProps, 'children'> {
    className?: string
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
}

export const Icon = React.forwardRef((props, reference) => {
    // TODO: role should have a default value of "img", but most of our Icons don't
    // provide an aria-label, title, or other form of alternative text, and so setting it
    // causes accessibility audits to fail in our integration test suite. Once we've added
    // text to all of our icons, we should restore this as the default value.
    // const { children, inline = true, className, size, as: Component = 'svg', role = 'img', ...attributes } = props
    const { children, inline = true, className, size, as: Component = 'svg', ...attributes } = props

    return (
        <Component
            className={classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className)}
            ref={reference}
            {...attributes}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<ComponentType<MdiReactIconProps> | ElementType, PropsWithChildren<IconProps>>

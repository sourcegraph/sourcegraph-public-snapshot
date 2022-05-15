import React, { ComponentType, ElementType, PropsWithChildren } from 'react'

import classNames from 'classnames'
import { MdiReactIconProps } from 'mdi-react'

import { ForwardReferenceComponent } from '../..'

import { ICON_SIZES } from './constants'

import styles from './Icon.module.scss'

interface BaseIconProps extends Omit<MdiReactIconProps, 'children'> {
    className?: string
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
}

interface ScreenReaderIconProps extends BaseIconProps {
    'aria-label'?: string
}

interface HiddenIconProps extends BaseIconProps {
    'aria-hidden'?: true | 'true'
}

// We're currently migrating our icons to provide a descriptive label or use aria-hidden to be excluded from screen readers.
// Migration issue: https://github.com/sourcegraph/sourcegraph/issues/34582
// TODO: We should enforce that these props are provided once that migration is complete.
export type IconProps = HiddenIconProps | ScreenReaderIconProps

export const Icon = React.forwardRef((props, reference) => {
    const { children, inline = true, className, size, as: Component = 'svg', role, ...attributes } = props

    return (
        <Component
            className={classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className)}
            ref={reference}
            role={role}
            {...attributes}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<
    ComponentType<React.PropsWithChildren<MdiReactIconProps>> | ElementType,
    PropsWithChildren<IconProps>
>

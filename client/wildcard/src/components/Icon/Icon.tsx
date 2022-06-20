import React, { AriaRole, ComponentType, ElementType } from 'react'

import MDIIcon from '@mdi/react'
import classNames from 'classnames'
import { MdiReactIconProps } from 'mdi-react'

import { ForwardReferenceComponent } from '../..'

import { ICON_SIZES } from './constants'

import styles from './Icon.module.scss'

interface BaseIconProps extends Omit<React.ComponentProps<typeof MDIIcon>, 'size' | 'path' | 'color'> {
    /**
     * Provide a custom `svgPath` to build an SVG.
     *
     * If using a Material Design icon, simply import the path from '@mdj/js'.
     */
    svgPath?: string
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
    className?: string
    role?: AriaRole
}

interface ScreenReaderIconProps extends BaseIconProps {
    'aria-label': string
}

interface HiddenIconProps extends BaseIconProps {
    'aria-hidden': true | 'true'
}

export type IconProps = HiddenIconProps | ScreenReaderIconProps

// eslint-disable-next-line react/display-name
export const Icon = React.forwardRef(({ children, className, size, ...props }, reference) => {
    const iconStyle = classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className)

    if (props.svgPath) {
        const { svgPath, 'aria-label': ariaLabel, ...attributes } = props

        return (
            <MDIIcon
                ref={reference as React.RefObject<SVGSVGElement>}
                path={svgPath}
                className={iconStyle}
                title={ariaLabel}
                {...attributes}
            />
        )
    }

    const { as: IconComponent = 'svg', role = 'img', ...attributes } = props

    return (
        <IconComponent ref={reference} className={iconStyle} role={role} {...attributes}>
            {children}
        </IconComponent>
    )
}) as ForwardReferenceComponent<ComponentType<React.PropsWithChildren<MdiReactIconProps>> | ElementType, IconProps>

Icon.displayName = 'Icon'

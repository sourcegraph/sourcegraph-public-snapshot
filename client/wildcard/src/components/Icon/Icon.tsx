import React, { type AriaRole, type ComponentType, type ElementType, type SVGProps } from 'react'

import classNames from 'classnames'
import type { MdiReactIconProps } from 'mdi-react'

import type { ForwardReferenceComponent } from '../..'

import type { ICON_SIZES } from './constants'

import styles from './Icon.module.scss'

type PathIcon = string
type CustomIcon = ComponentType<{ className?: string }>
export type IconType = PathIcon | CustomIcon

interface BaseIconProps extends SVGProps<SVGSVGElement> {
    /**
     * Provide a custom `svgPath` to build an SVG.
     *
     * If using a Material Design icon, simply import the path from '@mdj/js'.
     */
    svgPath?: PathIcon
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
    /**
     * If the icon should be styled to scale according to the surrounding text.
     * Defaults to `true`.
     */
    inline?: boolean
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

/**
 * Renders an Icon.
 *
 * We have two types of icons:
 * - [Material design icons](https://materialdesignicons.com): These should be imported from `@mdi/js` and used as the `svgPath` prop.
 * Example: `<Icon svgPath={mdiIcon} />`
 *
 * - Custom icons: We can use this directly by passing the icon component as the `as` prop.
 * Example: `<Icon as={SourcegraphLogo} />`
 *
 * Note: In order to be accessible, we enforce that either an `aria-label` OR an `aria-hidden` prop is provided.
 * If the icon is not decorative, and adds value to the users journey, we should use a descriptive `aria-label`.
 */
export const Icon = React.memo(
    React.forwardRef(function Icon(
        { children, className, size, role = 'img', inline = true, svgPath, ...props },
        reference
    ) {
        const iconStyle = classNames(
            'mdi-icon',
            inline && styles.iconInline,
            size === 'md' && styles.iconInlineMd,
            className
        )

        if (svgPath) {
            const { height = 24, width = 24, viewBox = '0 0 24 24', fill = 'currentColor', ...attributes } = props

            return (
                <svg
                    ref={reference}
                    className={iconStyle}
                    role={role}
                    height={height}
                    width={width}
                    viewBox={viewBox}
                    fill={fill}
                    {...attributes}
                >
                    <path d={svgPath} />
                </svg>
            )
        }

        const { as: IconComponent = 'svg', ...attributes } = props

        return (
            <IconComponent ref={reference} className={iconStyle} role={role} {...attributes}>
                {children}
            </IconComponent>
        )
    })
) as ForwardReferenceComponent<ComponentType<React.PropsWithChildren<MdiReactIconProps>> | ElementType, IconProps>

import MDIIcon from '@mdi/react'
import { MdiReactIconComponentType, MdiReactIconProps } from 'mdi-react'
import React from 'react'

import { AccessibleSvgProps, AccessibleSvg } from './AccessibleSvg'
import { IconStyle, IconStyleProps } from './IconStyle'

export type AccessibleIcon = typeof Icon | AccessibleSvg

interface BaseIconProps extends IconStyleProps {}

interface BasePathIconProps extends BaseIconProps, Omit<React.ComponentProps<typeof MDIIcon>, 'size' | 'path'> {
    /**
     * Provide a custom `svgPath` to build an SVG.
     *
     * If using a Material Design icon, simply import the path from '@mdj/js'.
     */
    svgPath: string
}
type PathIconProps = BasePathIconProps & AccessibleSvgProps

interface BaseComponentIconProps extends BaseIconProps, React.SVGAttributes<SVGElement> {
    /**
     * Provide a custom component to render an SVG.
     *
     * This should either:
     * - Be any React component that matches the `Icon` footprint (and exposes accessible props).
     * - Be a React component that wraps an SVG and implements the `AccessibleSvg` type.
     *
     * Note:  `mdi-react`
     */
    as: AccessibleIcon
}
/**
 * @deprecated Frontend Platform is phasing this out in favor of `@mdi/react`.
 */
interface LegacyComponentIconProps extends BaseIconProps, Omit<MdiReactIconProps, 'size'> {
    as: MdiReactIconComponentType
}
type ComponentIconProps = (BaseComponentIconProps & AccessibleSvgProps) | LegacyComponentIconProps

export type IconProps = PathIconProps | ComponentIconProps

export const Icon: React.FunctionComponent<IconProps> = ({ children, className, ...props }) => {
    if ('svgPath' in props) {
        const { svgPath, ...attributes } = props

        return <IconStyle as={MDIIcon} path={svgPath} className={className} {...attributes} />
    }

    const { as: IconComponent = 'svg', ...attributes } = props

    return <IconStyle as={IconComponent} className={className} {...attributes} />
}

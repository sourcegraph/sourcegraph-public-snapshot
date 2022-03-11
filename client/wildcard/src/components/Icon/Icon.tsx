import MDIIcon from '@mdi/react'
import React from 'react'

import { AccessibleSvg, AccessibleSvgProps } from './AccessibleSvgComponent'
import { IconStyle, IconStyleProps } from './IconStyle'

interface BaseIconProps extends IconStyleProps {}

interface PathIconProps extends BaseIconProps, Omit<React.ComponentProps<typeof MDIIcon>, 'size' | 'path'> {
    svgPath: string
}

interface ComponentIconProps extends BaseIconProps, React.SVGAttributes<SVGElement> {
    as: AccessibleSvg
}

export type IconProps = (PathIconProps & AccessibleSvgProps) | (ComponentIconProps & AccessibleSvgProps)

export const Icon: React.FunctionComponent<IconProps> = ({ children, className, ...props }) => {
    if ('svgPath' in props) {
        const { svgPath, ...attributes } = props

        return <IconStyle as={MDIIcon} path={svgPath} className={className} {...attributes} />
    }

    const { as: IconComponent = 'svg', ...attributes } = props

    return <IconStyle as={IconComponent} className={className} {...attributes} />
}

import MDIIcon from '@mdi/react'
import React from 'react'

import { AccessibleSVGComponent } from './AccessibleSvgComponent'
import { IconStyle, IconStyleProps } from './IconStyle'

interface BaseIconProps extends IconStyleProps {}
interface BasePathIconProps extends BaseIconProps, Omit<React.ComponentProps<typeof MDIIcon>, 'size' | 'path'> {
    svgPath: string
}

interface ScreenReaderPathIconProps extends BasePathIconProps {
    title: string
}
interface HiddenPathIconProps extends BasePathIconProps {
    'aria-hidden': true | 'true'
}
type PathIconProps = ScreenReaderPathIconProps | HiddenPathIconProps

interface ComponentIconProps extends BaseIconProps {
    as: AccessibleSVGComponent
}

export type IconProps = PathIconProps | ComponentIconProps

export const Icon: React.FunctionComponent<IconProps> = ({ children, className, ...props }) => {
    if ('svgPath' in props) {
        const { svgPath, ...attributes } = props

        return <IconStyle as={MDIIcon} path={svgPath} className={className} {...attributes} />
    }

    const { as: IconComponent = 'div', ...attributes } = props

    return <IconStyle as={IconComponent} className={className} {...attributes} />
}

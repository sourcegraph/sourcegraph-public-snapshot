import MDIIcon from '@mdi/react'
import React from 'react'

import { ForwardReferenceComponent } from '../..'

import { IconStyle, IconStyleProps } from './IconStyle'

interface BaseIconProps extends IconStyleProps {}

interface PathIconProps extends BaseIconProps, Omit<React.ComponentProps<typeof MDIIcon>, 'size' | 'path'> {
    svgPath: string
}

interface ComponentIconProps extends BaseIconProps {
    svg: React.ComponentType<{ className?: string }>
}

export type IconProps = PathIconProps | ComponentIconProps

export const Icon = React.forwardRef(({ children, className, ...props }, reference) => {
    if ('svgPath' in props) {
        const { svgPath, ...attributes } = props

        return (
            <IconStyle
                as={MDIIcon}
                path={svgPath}
                ref={reference && 'current' in reference ? reference : undefined}
                {...attributes}
            />
        )
    }

    const { svg, ...attributes } = props

    return <IconStyle as={svg} ref={reference} {...attributes} />
}) as ForwardReferenceComponent<'svg', IconProps>

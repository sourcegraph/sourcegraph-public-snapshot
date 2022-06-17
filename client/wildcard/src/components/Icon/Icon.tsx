import React, { ComponentType } from 'react'

import MDIIcon from '@mdi/react'

import { ForwardReferenceComponent } from '../..'

import { IconStyle, IconStyleProps } from './IconStyle'

interface BaseIconProps extends IconStyleProps, Omit<React.ComponentProps<typeof MDIIcon>, 'size' | 'path' | 'color'> {
    /**
     * Provide a custom `svgPath` to build an SVG.
     *
     * If using a Material Design icon, simply import the path from '@mdj/js'.
     */
    svgPath?: string
}

interface ScreenReaderIconProps extends BaseIconProps {
    'aria-label': string
}

interface HiddenIconProps extends BaseIconProps {
    'aria-hidden': true | 'true'
}

export type IconProps = HiddenIconProps | ScreenReaderIconProps

// eslint-disable-next-line react/display-name
export const IconV2 = React.forwardRef(({ children, className, ...props }, reference) => {
    if ('svgPath' in props && props.svgPath) {
        const { svgPath, ...attributes } = props

        return (
            <IconStyle
                as={MDIIcon}
                ref={reference as React.RefObject<SVGSVGElement>}
                path={svgPath}
                className={className}
                {...attributes}
            />
        )
    }

    const { as: IconComponent = 'svg', ...attributes } = props

    return <IconStyle as={IconComponent} ref={reference} className={className} {...attributes} />
}) as ForwardReferenceComponent<ComponentType | 'svg', IconProps>

IconV2.displayName = 'IconV2'

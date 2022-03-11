import React from 'react'

interface BaseSVGProps {
    className?: string
}

interface ScreenReaderIconProps extends BaseSVGProps {
    title: string
}

interface HiddenIconProps extends BaseSVGProps {
    'aria-hidden': true | 'true'
}

export type AccessibleSvgProps = ScreenReaderIconProps | HiddenIconProps

export type AccessibleSvg = React.ComponentType<AccessibleSvgProps & React.SVGAttributes<SVGElement>>
export type AccessibleIcon = React.ComponentType<AccessibleSvgProps>

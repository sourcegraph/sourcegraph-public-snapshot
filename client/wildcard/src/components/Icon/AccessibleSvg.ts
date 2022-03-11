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

export type AccessibleSvgProps = HiddenIconProps | ScreenReaderIconProps

/**
 * Accessible SVG
 *
 * This type should be used for components which purely wrap a raw <svg>.
 *
 * It will ensure that the component supports all possible SVG attributes, and enforces that the consumer provides suitable props for accessibility.
 *
 * Note: You should ensure that your SVG component makes correct use of the `title` or `aria-hidden` prop to ensure your graphic can be made accessible.
 */
export type AccessibleSvg = React.ComponentType<
    AccessibleSvgProps &
        React.SVGAttributes<SVGElement> & {
            /**
             * Shorthand for width=X and height=X
             */
            size?: number
        }
>

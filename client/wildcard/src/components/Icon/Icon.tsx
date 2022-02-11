import classNames from 'classnames'
import React from 'react'

import { ICON_SIZES } from './constants'
import styles from './Icon.module.scss'

interface IconProps {
    className?: string
    svg: React.SVGAttributes<SVGSVGElement>
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
    /**
     * Used to change the element that is rendered.
     * Always be mindful of potentially accessibiliy pitfalls when using this!
     */
    as?: React.ElementType
}

export const Icon: React.FunctionComponent<IconProps> = ({
    svg,
    className,
    size,
    as: Component = 'div',
    ...attributes
}) => (
    <Component
        className={classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className)}
        {...attributes}
    >
        {svg}
    </Component>
)

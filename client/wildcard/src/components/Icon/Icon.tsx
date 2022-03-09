import MDIIcon from '@mdi/react'
import classNames from 'classnames'
import React, { ElementType, SVGProps } from 'react'

import { ICON_SIZES } from './constants'
import styles from './Icon.module.scss'

interface BaseIconProps {
    className?: string
    /**
     * The variant style of the icon. defaults to 'sm'
     */
    size?: typeof ICON_SIZES[number]
}

interface LegacyIconProps extends BaseIconProps, SVGProps<SVGSVGElement> {
    as: ElementType
}

interface NewIconProps extends BaseIconProps, Omit<React.ComponentProps<typeof MDIIcon>, 'size'> {}

type IconProps = LegacyIconProps | NewIconProps

export const Icon: React.FunctionComponent<IconProps> = ({ children, className, size, ...props }) => {
    const sharedProps = {
        className: classNames(styles.iconInline, size === 'md' && styles.iconInlineMd, className),
    }

    if ('as' in props) {
        const { as: Component = 'svg', ...attributes } = props

        return (
            <Component {...sharedProps} {...attributes}>
                {children}
            </Component>
        )
    }

    const { path, ...attributes } = props

    return <MDIIcon {...sharedProps} {...attributes} path={path} />
}

// React.forwardRef(({ children, className, ...props }, reference) => (
//     <div ref={reference} {...props} className={classNames('dropdown-divider', className)} />
// )) as ForwardReferenceComponent<'div'>

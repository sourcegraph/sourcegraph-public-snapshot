import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'
import { getAlignmentStyle, getModeStyle, TypographyProps } from '../utils'

import styles from './Heading.module.scss'

export type HeadingProps = React.HTMLAttributes<HTMLHeadingElement> & TypographyProps
export type HeadingElement = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'

type InternalHeadingProps = HeadingProps & {
    styleAs?: HeadingElement
}

const getStyleAs = (headerX: HeadingElement | undefined): string | undefined =>
    headerX && styles[headerX as keyof typeof styles]

export const Heading = React.forwardRef(
    ({ children, as: Component = 'h1', styleAs = Component, alignment, mode, className, ...props }, reference) => (
        <Component
            className={classNames(
                getStyleAs(styleAs),
                className,
                alignment && getAlignmentStyle({ alignment }),
                mode && getModeStyle({ mode })
            )}
            {...props}
            ref={reference}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'h1' | HeadingElement, InternalHeadingProps>

import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'
import { getAlignmentStyle, getModeStyle, HeadingElement, TypographyProps } from '../utils'

import styles from './Heading.module.scss'

export type HeadingProps = React.HTMLAttributes<HTMLHeadingElement> &
    TypographyProps & {
        styleAs?: HeadingElement
    }

const getStyleAs = (headerX: HeadingElement | undefined): string | undefined =>
    headerX && styles[headerX as keyof typeof styles]

export const Heading = React.forwardRef(
    ({ children, as: Component = 'div', styleAs, alignment, mode, className, ...props }, reference) => (
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
) as ForwardReferenceComponent<'div', HeadingProps>

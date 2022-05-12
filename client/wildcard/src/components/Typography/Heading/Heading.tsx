import React from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../../types'
import { getAlignmentStyle, getModeStyle, TypographyProps } from '../utils'

export type HeadingProps = React.HTMLAttributes<HTMLHeadingElement> & TypographyProps

export const Heading = React.forwardRef(
    ({ children, as: Component = 'div', alignment, mode, className, ...props }, reference) => (
        <Component
            className={classNames(
                className,
                alignment && getAlignmentStyle({ alignment }),
                mode && getModeStyle({ mode })
            )}
            {...props}
            ref={reference}
            {...props}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', HeadingProps>

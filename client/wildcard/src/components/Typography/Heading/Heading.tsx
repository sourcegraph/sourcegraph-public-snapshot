import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../../types'
import { getAlignmentStyle, getModeStyle, TypographyProps } from '../utils'

export type HeadingProps = React.HTMLAttributes<HTMLHeadingElement> & TypographyProps

export const Heading = React.forwardRef(
    ({ children, as: Component = 'div', alignment, mode, className }, reference) => (
        <Component
            className={classNames(
                className,
                alignment && getAlignmentStyle({ alignment }),
                mode && getModeStyle({ mode })
            )}
            ref={reference}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'div', HeadingProps>

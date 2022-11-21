import React, { MouseEvent } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../../types'
import { TypographyProps } from '../utils'

import { getLabelClassName } from './utils'

interface LabelProps extends React.HTMLAttributes<HTMLLabelElement>, TypographyProps {
    size?: 'small' | 'base'
    weight?: 'regular' | 'medium' | 'bold'
    isUnderline?: boolean
    isUppercase?: boolean
}

export const Label = React.forwardRef((props, reference) => {
    const {
        children,
        as: Component = 'label',
        size,
        weight,
        alignment,
        mode,
        isUnderline,
        isUppercase,
        className,
        onClick,
        ...rest
    } = props

    const mergedRef = useMergeRefs([reference])

    // Listen clicks on label elements in order to improve click-to-focus logic
    // over contenteditable="true". Be default label element native behavior (click to focus first input element)
    // doesn't work with content editable element.
    // Since we use content editable inputs (codemirror search box input) and labels together in some
    // consumers we support this behavior manually for content editable elements
    const handleClick = (event: MouseEvent<HTMLLabelElement>): void => {
        const htmlForAttribute = mergedRef.current?.getAttribute('htmlFor')

        if (htmlForAttribute) {
            onClick?.(event)
            return
        }

        const inputElements = mergedRef.current?.querySelectorAll('input') ?? []

        if (inputElements.length === 0) {
            const contendEditableElement = mergedRef.current?.querySelector<HTMLElement>('[contenteditable="true"]')

            if (contendEditableElement) {
                // If there is content editable element focus it instead and prevent default behavior
                // We need to prevent default behavior because by default labels focuses any interactive elements
                // such buttons, in order to preserve focus on the content editable element we stop this behavior
                event.preventDefault()
                contendEditableElement.focus()
            }
        }

        onClick?.(event)
    }

    return (
        <Component
            ref={mergedRef}
            className={classNames(
                getLabelClassName({ isUppercase, isUnderline, alignment, weight, size, mode }),
                className
            )}
            onClick={handleClick}
            {...rest}
        >
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'label', LabelProps>

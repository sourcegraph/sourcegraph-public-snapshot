import React, { type ReactNode } from 'react'

import classNames from 'classnames'

import { InputDescription } from '../Input'
import type { AccessibleFieldProps } from '../internal/AccessibleFieldType'
import { FormFieldLabel } from '../internal/FormFieldLabel'
import { FormFieldMessage } from '../internal/FormFieldMessage'
import { getValidStyle } from '../internal/utils'

import styles from './Select.module.scss'

export const SELECT_SIZES = ['sm', 'lg'] as const

export type SelectProps = AccessibleFieldProps<React.SelectHTMLAttributes<HTMLSelectElement>> &
    React.RefAttributes<HTMLSelectElement> & {
        /** Description block shown above the input (but below the label) */
        description?: ReactNode
        /**
         * Use the global `custom-select` class.
         */
        isCustomStyle?: boolean
        /**
         * Optional size modifier to render a smaller or larger <select> variant
         */
        selectSize?: typeof SELECT_SIZES[number]
        /**
         * Custom class name for select element.
         */
        selectClassName?: string
        /**
         * Optional label position. Default is 'inline'
         */
        labelVariant?: 'inline' | 'block'
        /**
         * Custom class name for label element.
         */
        labelClassName?: string
    }

/**
 * Returns the global CSS class to differentiate between native and custom <select> styles.
 */
export const getSelectStyles = ({
    isCustomStyle,
    selectSize,
}: Pick<SelectProps, 'isCustomStyle' | 'selectSize'>): string => {
    if (isCustomStyle) {
        return classNames('custom-select', selectSize && `custom-select-${selectSize}`)
    }

    return classNames('form-control', selectSize && `form-control-${selectSize}`)
}

/**
 * A wrapper around the <select> element.
 * Supports both native and custom styling.
 *
 * Select should be used to provide a user with a list of options within a form.
 *
 * Please note that this component takes <option> elements as children. This is to easily support advanced functionality such as usage of <optgroup>.
 */
export const Select: React.FunctionComponent<React.PropsWithChildren<SelectProps>> = React.forwardRef(function Select(
    {
        children,
        className,
        selectClassName,
        labelClassName,
        message,
        isValid,
        isCustomStyle,
        selectSize,
        labelVariant = 'inline',
        description,
        ...props
    },
    reference
) {
    return (
        <div className={classNames('form-group', className)}>
            {'label' in props && (
                <FormFieldLabel
                    htmlFor={props.id}
                    className={classNames(labelVariant === 'block' && styles.labelBlock, labelClassName)}
                >
                    {props.label}
                </FormFieldLabel>
            )}
            {description && <InputDescription className="ml-0 mb-2 mt-n1">{description}</InputDescription>}
            {/* eslint-disable-next-line react/forbid-elements */}
            <select
                ref={reference}
                className={classNames(
                    getSelectStyles({ isCustomStyle, selectSize }),
                    getValidStyle(isValid),
                    selectClassName
                )}
                {...props}
            >
                {children}
            </select>
            {message && <FormFieldMessage isValid={isValid}>{message}</FormFieldMessage>}
        </div>
    )
})

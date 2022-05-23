import { forwardRef, ForwardRefExoticComponent, ReactNode, RefAttributes, TextareaHTMLAttributes } from 'react'

import classNames from 'classnames'

import { Input } from '../Input'
import { getValidStyle } from '../internal/utils'

import styles from './TextArea.module.scss'

export interface TextAreaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
    /** Title of textarea. Used as label */
    label?: ReactNode
    /** Description block shown below the textarea. */
    message?: ReactNode
    /** Custom class name for root label element. */
    className?: string
    /**
     * Used to control the styling of the field and surrounding elements.
     * Set this value to `false` to show invalid styling.
     * Set this value to `true` to show valid styling.
     */
    isValid?: boolean
    /** Disable textarea behavior */
    disabled?: boolean
    /**
     * Allow resizing
     *
     * @default true
     */
    resizable?: boolean
    /** Determines the size of the textarea */
    size?: 'regular' | 'small'
    /** Custom class name for textarea element. */
    inputClassName?: string
}

/**
 * Displays a textarea with description, error message, visual invalid and valid states.
 */
export const TextArea: ForwardRefExoticComponent<TextAreaProps & RefAttributes<HTMLTextAreaElement>> = forwardRef(
    (props, reference) => {
        const {
            label,
            message,
            className,
            disabled,
            isValid,
            size,
            inputClassName,
            resizable = true,
            ...otherProps
        } = props

        return (
            <Input.Label label={label} className={classNames(styles.label, className)}>
                {/* eslint-disable-next-line react/forbid-elements */}
                <textarea
                    disabled={disabled}
                    className={classNames(
                        styles.textarea,
                        'form-control',
                        getValidStyle(isValid),
                        size === 'small' && 'form-control-sm',
                        resizable === false && styles.resizeNone,
                        inputClassName
                    )}
                    {...otherProps}
                    ref={reference}
                />
                {message && <Input.Message isValid={isValid}>{message}</Input.Message>}
            </Input.Label>
        )
    }
)

TextArea.displayName = 'TextArea'

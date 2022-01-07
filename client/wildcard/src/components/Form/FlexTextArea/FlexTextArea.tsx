import classNames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, Ref, useEffect, useImperativeHandle, useRef, useState } from 'react'

import { TextArea } from '@sourcegraph/wildcard'

import { TextAreaProps } from '../TextArea'

import styles from './FlexTextArea.module.scss'

/**
 * Sync value of line height with our global styles
 * See line-height-base scss variable with 20/14 value.
 */
const DEFAULT_TEXTAREA_LINE_HEIGHT = 20

export type IProps = {
    initialRow?: number
    minRows?: number
    maxRows?: number
} & InputHTMLAttributes<HTMLInputElement | HTMLTextAreaElement>

/**
 * Flexible and auto-growable textarea element.
 *
 * This component is using textarea as the input component, but in order to support a passing this
 * component to the combobox input component that can take only HTMLInputElement we have to
 * cast all public props of this component from textarea to input element props.
 */
export const FlexTextArea = forwardRef(
    (props: IProps, reference: Ref<HTMLInputElement | HTMLTextAreaElement | null>) => {
        const { initialRow = 1, minRows = 1, maxRows = Infinity, className, value, ...otherProps } = props
        const [rows, setRows] = useState(initialRow)
        const innerReference = useRef<HTMLTextAreaElement>(null)

        // Casting ref from textarea to input element for top level (consumer) ref support
        useImperativeHandle(reference, () => (innerReference.current as unknown) as HTMLInputElement)

        useEffect(() => {
            const target = innerReference.current

            if (!target) {
                return
            }

            const previousRows = target.rows
            const textareaLineHeight = parseFloat(getComputedStyle(target).lineHeight) ?? DEFAULT_TEXTAREA_LINE_HEIGHT

            // reset number of rows in textarea
            target.rows = minRows

            const currentRows = Math.floor(target.scrollHeight / textareaLineHeight)

            if (currentRows === previousRows) {
                target.rows = currentRows
            }

            if (currentRows > previousRows) {
                target.scrollTop = target.scrollHeight
            }

            setRows(currentRows < maxRows ? currentRows : maxRows)
        }, [maxRows, minRows, value])

        return (
            <TextArea
                {...(otherProps as TextAreaProps)}
                value={value}
                ref={innerReference}
                rows={rows}
                className={classNames(styles.textarea, className)}
            />
        )
    }
)

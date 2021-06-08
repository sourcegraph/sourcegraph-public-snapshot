import classNames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, Ref, useEffect, useImperativeHandle, useRef, useState } from 'react'

import styles from './FlexTextarea.module.scss'

export type IProps = {
    initialRow?: number
    minRows?: number
    maxRows?: number
} & InputHTMLAttributes<HTMLInputElement>

/**
 * Flexible and auto-growable textarea element.
 *
 * This component is using textarea as the input component, but in order to support a passing this
 * component to the combobox input component that can take only HTMLInputElement we have to
 * cast all public props of this component from textarea to input element props.
 */
export const FlexTextArea = forwardRef((props: IProps, reference: Ref<HTMLInputElement | null>) => {
    const { initialRow = 1, minRows = 1, maxRows = Infinity, className, value, ...otherProps } = props
    const [rows, setRows] = useState(initialRow)
    const innerReference = useRef<HTMLTextAreaElement>(null)

    // Casting ref from textarea to input element for top level (consumer) ref support
    useImperativeHandle(reference, () => (innerReference.current as unknown) as HTMLInputElement)

    useEffect(() => {
        const textareaLineHeight = 22
        const target = innerReference.current

        if (!target) {
            return
        }

        const previousRows = target.rows
        target.rows = minRows // reset number of rows in textarea

        const currentRows = Math.floor(target.scrollHeight / textareaLineHeight)

        if (currentRows === previousRows) {
            target.rows = currentRows
        }

        if (currentRows > previousRows) {
            target.scrollTop = target.scrollHeight
        }

        setRows(currentRows < maxRows ? currentRows : maxRows)
    }, [maxRows, minRows, value])

    const classes = classNames(styles.textarea, className)

    return (
        <textarea
            {...(otherProps as InputHTMLAttributes<HTMLTextAreaElement>)}
            value={value}
            ref={innerReference}
            rows={rows}
            className={classes}
        />
    )
})

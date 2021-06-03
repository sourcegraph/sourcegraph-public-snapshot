import classNames from 'classnames'
import React, {
    forwardRef, InputHTMLAttributes, Ref,
    useEffect,
    useImperativeHandle,
    useRef,
    useState,
} from 'react'

import styles from './FlexTextarea.module.scss'

export type IProps = {
    initialRow?: number
    minRows?: number
    maxRows?: number
} & InputHTMLAttributes<HTMLInputElement>

/**
 * Flexible and auto-growable textarea element.
 */
export const FlexTextarea = forwardRef((props: IProps, reference: Ref<HTMLInputElement | null>) => {
    const {
        initialRow = 1,
        minRows = 1,
        maxRows = Infinity,
        className,
        value,
        ...otherProps } = props
    const [rows, setRows] = useState(initialRow)
    const innerReference = useRef<HTMLTextAreaElement>(null)

    useImperativeHandle(reference, () => innerReference.current as unknown as HTMLInputElement)

    useEffect(() => {
        const textareaLineHeight = 22
        const target = innerReference.current

        if (!target) {
            return
        }

        const previousRows = target.rows
        target.rows = minRows // reset number of rows in textarea

        const currentRows = ~~(target.scrollHeight / textareaLineHeight)

        if (currentRows === previousRows) {
            target.rows = currentRows
        }

        if (currentRows > previousRows) {
            target.scrollTop = target.scrollHeight
        }

        setRows(currentRows < maxRows ? currentRows : maxRows)
    }, [maxRows, minRows, value])

    const classes = classNames(
        styles.textarea,
        className,
    )

    return (
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        <textarea
            {...otherProps}
            value={value}
            ref={innerReference}
            rows={rows}
            className={classes} />
    )
})


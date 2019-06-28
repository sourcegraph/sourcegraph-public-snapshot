import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Form } from '../../../components/Form'

interface Props {
    /** The current value of the filter. */
    value: string

    /** Called when the value changes. */
    onChange: (value: string) => void

    beforeInputFragment?: React.ReactFragment

    className?: string
}

/**
 * The filter control (dropdown and input field) for a list of tasks.
 */
// tslint:disable: jsx-no-lambda
export const TasksListFilter: React.FunctionComponent<Props> = ({
    value,
    onChange,
    beforeInputFragment,
    className,
}) => {
    const [uncommittedValue, setUncommittedValue] = useState(value)
    useEffect(() => setUncommittedValue(value), [value])

    const [isFocused, setIsFocused] = useState(false)
    const onFocus = useCallback(() => setIsFocused(true), [])
    const onBlur = useCallback(() => setIsFocused(false), [])

    const prependSearchIcon = !beforeInputFragment

    return (
        <Form
            className={`form ${className}`}
            onSubmit={e => {
                e.preventDefault()
                onChange(uncommittedValue)
            }}
        >
            <div
                className={`input-group ${prependSearchIcon ? 'bg-form-control border rounded' : ''} ${
                    isFocused ? 'form-control-focus' : ''
                }`}
            >
                {beforeInputFragment || (
                    <div className="input-group-prepend">
                        <span className="input-group-text border-0 pl-2 pr-1 bg-transparent">
                            <SearchIcon className="icon-inline" />
                        </span>
                    </div>
                )}
                <input
                    type="text"
                    className={`form-control ${prependSearchIcon ? 'shadow-none border-0 rounded-0 pl-1' : ''}`}
                    aria-label="Filter tasks"
                    autoCapitalize="off"
                    value={uncommittedValue}
                    onChange={e => setUncommittedValue(e.currentTarget.value)}
                    onFocus={onFocus}
                    onBlur={onBlur}
                />
            </div>
        </Form>
    )
}

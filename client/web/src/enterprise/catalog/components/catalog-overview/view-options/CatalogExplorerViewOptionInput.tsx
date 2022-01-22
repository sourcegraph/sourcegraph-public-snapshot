import classNames from 'classnames'
import React, { useCallback } from 'react'

interface Props<T extends string> {
    label: string
    values: T[]
    selected: T | undefined
    onChange: (value: T | undefined) => void
    className?: string
}

export const CatalogExplorerViewOptionInput = <T extends string>({
    label,
    values,
    selected,
    onChange,
    className,
}: Props<T>): React.ReactElement<Props<T>> => {
    const innerOnChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            onChange((event.target.value || undefined) as T | undefined)
        },
        [onChange]
    )

    return (
        <>
            <select className={classNames('form-control', className)} value={selected} onChange={innerOnChange}>
                <option value="">{label}</option>
                {values.map(value => (
                    <option value={value} key={value}>
                        {value}
                    </option>
                ))}
            </select>
        </>
    )
}

import React from 'react'

import classNames from 'classnames'

import { Select } from '@sourcegraph/wildcard'

interface HorizontalSelect<T> {
    onChange: (value: T) => void
    value: T
    label: string
    className?: string
    items: { label: string; value: T; disabled?: boolean }[]
}

export const HorizontalSelect = <T extends string>({
    items,
    label,
    value,
    onChange,
    className,
}: React.PropsWithChildren<HorizontalSelect<T>>): JSX.Element => (
    <Select
        id="date-range"
        label={label}
        isCustomStyle={true}
        className={classNames('d-flex align-items-center m-0', className)}
        labelClassName="mb-0"
        selectClassName="ml-2"
        value={value}
        onChange={value => onChange(value.target.value as T)}
    >
        {items.map(({ value, label, disabled }) => (
            <option key={label} value={value} disabled={disabled}>
                {label}
            </option>
        ))}
    </Select>
)

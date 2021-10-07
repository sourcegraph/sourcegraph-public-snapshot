import React, { FunctionComponent } from 'react'

import { defaultDurationValues } from './shared'

export interface DurationSelectProps {
    id: string
    value: string | null
    disabled: boolean
    onChange?: (value: number | null) => void
    durationValues?: { value: number; displayText: string }[]
}

export const DurationSelect: FunctionComponent<DurationSelectProps> = ({
    id,
    value,
    disabled,
    onChange,
    durationValues = defaultDurationValues,
}) => (
    <select
        id={id}
        className="form-control"
        value={value || undefined}
        disabled={disabled}
        onChange={event => onChange?.(!event.target.value ? null : Math.floor(parseInt(event.target.value, 10)))}
    >
        {durationValues.map(({ value, displayText }) => (
            <option key={value} value={value || undefined}>
                {displayText}
            </option>
        ))}
    </select>
)

import { FunctionComponent, useState } from 'react'

import classNames from 'classnames'

import { Select, Input } from '@sourcegraph/wildcard'

import { defaultDurationValues } from '../shared'

export interface DurationSelectProps {
    id: string
    value: string | null
    disabled: boolean
    onChange?: (value: number | null) => void
    durationValues?: { value: number; displayText: string }[]
    className?: string
}

const defaultCustomValue = 24

const toInt = (value: string): number | null => Math.floor(parseInt(value, 10)) || null

export const DurationSelect: FunctionComponent<React.PropsWithChildren<DurationSelectProps>> = ({
    id,
    value,
    disabled,
    onChange,
    durationValues = defaultDurationValues,
    className,
}) => {
    const [isCustom, setIsCustom] = useState(
        value !== null && !durationValues.map(({ value }) => value).includes(toInt(value))
    )

    return (
        <div className="input-group">
            <Select
                aria-label=""
                id={id}
                className={classNames('flex-1 mb-0', className)}
                value={isCustom ? 'custom' : value || undefined}
                disabled={disabled}
                onChange={event => {
                    if (event.target.value === 'custom') {
                        setIsCustom(true)
                    } else {
                        setIsCustom(false)
                        onChange?.(toInt(event.target.value))
                    }
                }}
            >
                {durationValues.map(({ value, displayText }) => (
                    <option key={value} value={value || undefined}>
                        {displayText}
                    </option>
                ))}

                <option value="custom">Custom</option>
            </Select>

            {isCustom && (
                <>
                    <Input
                        type="number"
                        className="ml-2"
                        value={value || defaultCustomValue}
                        min="1"
                        max="219150"
                        disabled={disabled}
                        onChange={event => onChange?.(toInt(event.target.value))}
                    />

                    <div className="input-group-append">
                        <span className="input-group-text"> hours </span>
                    </div>
                </>
            )}
        </div>
    )
}

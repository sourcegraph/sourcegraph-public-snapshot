import React from 'react'

import { ButtonGroup, Button, Tooltip } from '@sourcegraph/wildcard'

interface ToggleSelectProps<T> {
    selected: T
    className?: string
    items: {
        tooltip: string
        label: string
        value: T
    }[]
    onChange: (value: T) => void
}

export const ToggleSelect = <T extends any>({
    selected,
    items,
    onChange,
    className,
}: React.PropsWithChildren<ToggleSelectProps<T>>): JSX.Element => (
    <ButtonGroup className={className}>
        {items.map(({ tooltip, label, value }) => (
            <Tooltip key={label} content={tooltip} placement="top">
                <Button
                    onClick={() => onChange(value)}
                    outline={selected !== value}
                    variant={selected !== value ? 'secondary' : 'primary'}
                    display="inline"
                    size="sm"
                >
                    {label}
                </Button>
            </Tooltip>
        ))}
    </ButtonGroup>
)

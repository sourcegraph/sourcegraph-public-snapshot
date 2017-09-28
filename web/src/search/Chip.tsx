import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'

export interface ChipProps {
    icon: React.ComponentType
    label: string
    onDelete: () => void
}

export const Chip = (props: ChipProps) => (
    <span className='chip'>
        <props.icon />
        <span className='chip__label'>{props.label}</span>
        <button type='button' className='chip__delete-button' onClick={props.onDelete}>
            <CloseIcon />
        </button>
    </span>
)

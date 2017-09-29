import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'

export interface ChipProps {
    icon: React.ComponentType<{ className: string }>
    label: string
    onDelete: () => void
}

export const Chip = (props: ChipProps) => (
    <span className='chip'>
        <props.icon className='icon-inline'/>
        <span className='chip__label'>{props.label}</span>
        <button type='button' className='chip__delete-button' onClick={props.onDelete}>
            <CloseIcon className='icon-inline' />
        </button>
    </span>
)

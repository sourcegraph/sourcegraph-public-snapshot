import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'

export interface Props {
    title: string
    onDismiss: () => void
}

export const TreeHeader = (props: Props) => (
    <div className="tree-header">
        <span className="tree-header__title">{props.title}</span>
        <button onClick={props.onDismiss} className="btn btn-icon tree-header__close-button">
            <CloseIcon />
        </button>
    </div>
)

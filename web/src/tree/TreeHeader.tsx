
import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'

export interface Props {
    title: string
    onDismiss: () => void
}

export class TreeHeader extends React.Component<Props, {}> {
    public render(): JSX.Element {
        return (
            <div className='tree-header'>
                <span className='tree-header__title'>{this.props.title}</span>
                <button onClick={this.props.onDismiss} className='btn btn-icon tree-header__close-button'>
                    <CloseIcon />
                </button>
            </div>
        )
    }
}

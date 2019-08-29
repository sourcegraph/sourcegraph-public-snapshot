import * as React from 'react'

interface Props {
    value: string
}

export class DotStarButton extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <button
                type="button"
                className="btn btn-sm text-nowrap dot-star-button"
                data-testid="dot-star-button"
                value={this.props.value}
                title="Regular expression search style"
                onMouseDown={this.onMouseDown}
                onClick={this.onClick}
            >
                <div>.*</div>
            </button>
        )
    }

    private onMouseDown: React.MouseEventHandler<HTMLButtonElement> = event => {
        // Prevent clicking on the .* button from taking focus away from the search input.
        event.preventDefault()
    }

    private onClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        event.preventDefault()
    }
}

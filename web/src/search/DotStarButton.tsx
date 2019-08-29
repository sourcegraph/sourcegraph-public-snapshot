import * as React from 'react'

interface Props {}

export class DotStarButton extends React.PureComponent<Props> {
    constructor(props: Props) {
        super(props)
        this.state = {
            enabled: false,
        }
    }

    public render(): JSX.Element | null {
        return (
            <button
                type="button"
                className={`btn btn-sm text-nowrap dot-star-button ${
                    this.state.enabled ? ' dot-star-button--selected' : ''
                }`}
                data-testid="dot-star-button"
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
        this.setState({ enabled: !this.state.enabled })
    }
}

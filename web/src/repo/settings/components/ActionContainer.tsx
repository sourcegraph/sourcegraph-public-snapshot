import * as React from 'react'

export const BaseActionContainer: React.SFC<{
    title: React.ReactFragment
    description: React.ReactFragment
    action: React.ReactFragment
    details?: React.ReactFragment
}> = ({ title, description, action, details }) => (
    <div className="action-container">
        <div className="action-container__row">
            <div className="action-container__description">
                <h4 className="action-container__title">{title}</h4>
                {description}
            </div>
            <div className="action-container__btn-container">{action}</div>
        </div>
        {details && <div className="action-container__row">{details}</div>}
    </div>
)

interface Props {
    title: React.ReactFragment
    description: React.ReactFragment
    buttonLabel: React.ReactFragment

    /** The message to briefly display below the button when the action is successful. */
    flashText?: string

    run: () => Promise<void>
}

interface State {
    loading: boolean
    flash: boolean
    error?: string
}

/**
 * Displays an action button in a container with a title and description.
 */
export class ActionContainer extends React.PureComponent<Props, State> {
    public state: State = {
        loading: false,
        flash: false,
    }

    public render(): JSX.Element | null {
        return (
            <BaseActionContainer
                title={this.props.title}
                description={this.props.description}
                action={[
                    <button
                        key={0}
                        className="btn btn-primary action-container__btn"
                        onClick={this.onClick}
                        disabled={this.state.loading}
                    >
                        {this.props.buttonLabel}
                    </button>,
                    this.props.flashText && (
                        <div
                            key={1}
                            className={
                                'action-container__flash' +
                                (this.state.flash ? ' action-container__flash--visible' : '')
                            }
                        >
                            <small>{this.props.flashText}</small>
                        </div>
                    ),
                ]}
            />
        )
    }

    private onClick = () => {
        this.setState({
            error: undefined,
            loading: true,
        })

        this.props.run().then(
            () => {
                this.setState({ loading: false, flash: true })
                setTimeout(() => this.setState({ flash: false }), 1000)
            },
            err => this.setState({ loading: false, error: err.message })
        )
    }
}

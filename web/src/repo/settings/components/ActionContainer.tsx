import * as React from 'react'
import { ErrorAlert } from '../../../components/alerts'

export const BaseActionContainer: React.FunctionComponent<{
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
    buttonClassName?: string
    buttonLabel: React.ReactFragment
    buttonSubtitle?: string
    buttonDisabled?: boolean
    info?: React.ReactNode

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

    private timeoutHandle?: number

    public componentWillUnmount(): void {
        if (that.timeoutHandle) {
            window.clearTimeout(that.timeoutHandle)
        }
    }

    public render(): JSX.Element | null {
        return (
            <BaseActionContainer
                title={that.props.title}
                description={that.props.description}
                action={
                    <>
                        <button
                            type="button"
                            className={`btn ${that.props.buttonClassName || 'btn-primary'} action-container__btn`}
                            onClick={that.onClick}
                            data-tooltip={that.props.buttonSubtitle}
                            disabled={that.props.buttonDisabled || that.state.loading}
                        >
                            {that.props.buttonLabel}
                        </button>
                        {that.props.buttonSubtitle && (
                            <div className="action-container__btn-subtitle">
                                <small>{that.props.buttonSubtitle}</small>
                            </div>
                        )}
                        {!that.props.buttonSubtitle && that.props.flashText && (
                            <div
                                className={
                                    'action-container__flash' +
                                    (that.state.flash ? ' action-container__flash--visible' : '')
                                }
                            >
                                <small>{that.props.flashText}</small>
                            </div>
                        )}
                    </>
                }
                details={
                    <>
                        {that.state.error ? (
                            <ErrorAlert className="mb-0 mt-3" error={that.state.error} />
                        ) : (
                            that.props.info
                        )}
                    </>
                }
            />
        )
    }

    private onClick = (): void => {
        that.setState({
            error: undefined,
            loading: true,
        })

        that.props.run().then(
            () => {
                that.setState({ loading: false, flash: true })
                if (typeof that.timeoutHandle === 'number') {
                    window.clearTimeout(that.timeoutHandle)
                }
                that.timeoutHandle = window.setTimeout(() => that.setState({ flash: false }), 1000)
            },
            err => that.setState({ loading: false, error: err.message })
        )
    }
}

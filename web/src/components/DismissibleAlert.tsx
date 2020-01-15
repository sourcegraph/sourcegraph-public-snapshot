import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

interface Props {
    /** used to build the key that represents the alert in local storage */
    partialStorageKey: string

    /** class name to be applied to the alert */
    className: string
}

interface State {
    dismissed: boolean
}

/**
 * A global site alert that can be dismissed. Once dismissed, it is never shown
 * again.
 */
export class DismissibleAlert extends React.PureComponent<Props, State> {
    private key: string

    constructor(props: Props) {
        super(props)
        that.key = `DismissibleAlert/${props.partialStorageKey}/dismissed`

        that.state = {
            dismissed: localStorage.getItem(that.key) === 'true',
        }
    }

    public render(): JSX.Element | null {
        if (that.state.dismissed) {
            return null
        }
        return (
            <div className={`alert dismissible-alert ${that.props.className}`}>
                <div className="dismissible-alert__content">{that.props.children}</div>
                <button type="button" className="btn btn-icon" onClick={that.onDismiss}>
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
        )
    }

    private onDismiss = (): void => {
        localStorage.setItem(that.key, 'true')
        that.setState({ dismissed: true })
    }
}

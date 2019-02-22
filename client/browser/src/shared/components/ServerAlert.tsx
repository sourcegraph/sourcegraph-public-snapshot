import CloseIcon from 'mdi-react/CloseIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import * as React from 'react'
import storage from '../../browser/storage'

interface Props {
    onClose: () => void
}

export const SERVER_CONFIGURATION_KEY = 'NeedsServerConfigurationAlertDismissed'

/**
 * A global alert telling the user that they need to configure Sourcegraph
 * to get code intelligence and search on private code.
 */
export class NeedsServerConfigurationAlert extends React.Component<Props, {}> {
    private sync(): void {
        storage.setSync({ [SERVER_CONFIGURATION_KEY]: true }, () => {
            this.props.onClose()
        })
    }

    private onClicked = () => {
        this.sync()
    }

    private onClose = () => {
        this.sync()
    }

    public render(): JSX.Element | null {
        return (
            <div className="sg-alert sg-alert-warning site-alert" style={{ justifyContent: 'space-between' }}>
                <a
                    onClick={this.onClicked}
                    className="site-alert__link"
                    href="https://docs.sourcegraph.com"
                    target="_blank"
                >
                    <span className="icon-inline site-alert__link-icon">
                        <WarningIcon size={17} />
                    </span>{' '}
                    <span className="underline">Configure Sourcegraph</span>
                </a>
                &nbsp;for code intelligence on private repositories.
                <div
                    style={{
                        display: 'inline-flex',
                        flex: '1 1 auto',
                        textAlign: 'right',
                        flexDirection: 'row-reverse',
                    }}
                >
                    <span
                        onClick={this.onClose}
                        style={{
                            fill: 'white',
                            cursor: 'pointer',
                            width: 17,
                            height: 17,
                            color: 'white',
                            paddingTop: 3,
                        }}
                    >
                        <CloseIcon size={17} />
                    </span>
                </div>
            </div>
        )
    }
}

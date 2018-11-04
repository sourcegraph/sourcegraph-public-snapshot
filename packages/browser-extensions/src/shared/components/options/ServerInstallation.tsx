import React from 'react'
import { Button } from 'reactstrap'
import storage from '../../../browser/storage'
import { isSourcegraphDotCom } from '../../util/context'
import {
    LEARN_MORE_URL,
    QUICK_START_URL,
    SERVER_INSTALLATION_BODY,
    SERVER_INSTALLATION_HEADER,
    SERVER_INSTALLATION_LINK_CTA,
    SERVER_INSTALLATION_PRIMARY_CTA,
} from './utils'

interface State {
    showSection: boolean
}

export class ServerInstallation extends React.Component<{}, State> {
    public state = {
        showSection: false,
    }

    public componentDidMount(): void {
        storage.getSync(items => {
            this.setState(() => ({ showSection: isSourcegraphDotCom(items.sourcegraphURL) }))
        })
    }

    private openQuickStart = () => {
        window.open(QUICK_START_URL, '_blank')
    }

    private openLearnMore = () => {
        window.open(LEARN_MORE_URL, '_blank')
    }

    public render(): JSX.Element | null {
        if (!this.state.showSection) {
            return null
        }

        return (
            <div>
                <div className="options__divider" />
                <div className="options__section">
                    <div className="options__section-header">{SERVER_INSTALLATION_HEADER}</div>
                    <div className="options__section-contents">
                        {SERVER_INSTALLATION_BODY}
                        <div className="options__installation-cta">
                            <Button className="options__cta" color="primary" onClick={this.openQuickStart}>
                                {SERVER_INSTALLATION_PRIMARY_CTA}
                            </Button>{' '}
                            <Button className="options__link" color="link" onClick={this.openLearnMore}>
                                {SERVER_INSTALLATION_LINK_CTA}
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}

import * as React from 'react'
import { Redirect } from 'react-router-dom'

export class LicenseInvalidPage extends React.Component<{}, {}> {
    constructor(props: {}) {
        super(props)
    }

    public render(): JSX.Element | null {
        if (window.context.licenseStatus === 'valid') {
            return <Redirect to="/search" />
        }
        return (
            <div className="license-page">
                <div className="warning-sign">&#9888;</div>
                {this.renderInternal()}
            </div>
        )
    }

    public renderInternal(): JSX.Element {
        switch (window.context.licenseStatus) {
            case 'expired':
                return (
                    <div>
                        Your license is expired. Please contact <MailToLink /> to obtain or renew your license.
                    </div>
                )
            case 'missing':
                return (
                    <div>
                        Your license is missing. Please contact <MailToLink /> to obtain a license.
                    </div>
                )
            case 'invalid':
            default:
                return (
                    <div>
                        Your license was invalid. Please contact <MailToLink /> to obtain a valid license.
                    </div>
                )
        }
    }
}

class MailToLink extends React.Component<{}, any> {
    constructor(props: {}) {
        super(props)
    }

    public render(): JSX.Element {
        const address = 'sales@sourcegraph.com'
        const encodedSubject = encodeURIComponent('License Request')
        const encodedBody = encodeURIComponent(
            'I would like to request a Sourcegraph Server license for my organization, __________'
        )
        return <a href={'mailto:' + address + '?subject=' + encodedSubject + '&body=' + encodedBody}>{address}</a>
    }
}

import { Warning } from '@sourcegraph/icons/lib/Warning'
import * as React from 'react'
import { Redirect } from 'react-router-dom'
import { PageTitle } from './components/PageTitle'

const MailToLink = () => {
    const address = 'sales@sourcegraph.com'
    const p = new URLSearchParams()
    p.set('subject', 'License Request')
    p.set('body', 'I would like to request a Sourcegraph Server license for my organization, __________')
    return <a href={'mailto:' + address + '?' + p.toString()}>{address}</a>
}

export class LicenseInvalidPage extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        if (window.context.licenseStatus === 'valid') {
            return <Redirect to="/search" />
        }
        return (
            <div className="license-invalid-page">
                <PageTitle title="License unverified" />
                <div className="license-invalid-page__warning-sign">
                    <Warning />
                </div>
                {(() => {
                    switch (window.context.licenseStatus) {
                        case 'expired':
                            return (
                                <div>
                                    Your license is expired. Please contact <MailToLink /> to obtain or renew your
                                    license.
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
                                    Your license is invalid. Please contact <MailToLink /> to obtain a valid license.
                                </div>
                            )
                    }
                })()}
            </div>
        )
    }
}

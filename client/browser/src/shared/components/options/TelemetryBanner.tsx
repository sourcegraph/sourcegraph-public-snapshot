import * as React from 'react'
import { Alert } from 'reactstrap'

export const TelemetryBanner: React.SFC<{}> = () => (
    <div className="options__alert">
        <Alert className="options__alert-warning">
            No private code, private repository names, personal data, or usage data is ever sent to Sourcegraph.com.
            {'\n'}
            <a
                href="https://about.sourcegraph.com/privacy/"
                target="_blank"
                // tslint:disable-next-line
                onClick={e => e.stopPropagation()}
                className="options__alert-link"
            >
                (Privacy Policy)
            </a>
        </Alert>
    </div>
)

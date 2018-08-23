import * as React from 'react'
import { Alert } from 'reactstrap'

export const ConfigWarning: React.SFC<{}> = () => (
    <div className="options__alert">
        <Alert className="options__alert-warning">
            Warning: changing the below options may break your browser extension.{' '}
            <a href="https://about.sourcegraph.com/docs/server/" className="options__alert-link" target="_blank">
                Learn More
            </a>
            .
        </Alert>
    </div>
)

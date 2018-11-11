import * as React from 'react'
import { Alert } from 'reactstrap'

export const ConfigWarning: React.SFC<{}> = () => (
    <div className="options__alert">
        <Alert className="options__alert-warning">
            Warning: changing the below options may break your browser extension.{' '}
            <a href="https://docs.sourcegraph.com" className="options__alert-link" target="_blank">
                Learn more.
            </a>
        </Alert>
    </div>
)

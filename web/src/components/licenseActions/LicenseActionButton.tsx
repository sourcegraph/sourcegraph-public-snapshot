import React from 'react'

interface Props {}

let OverrideComponent: React.FunctionComponent<Props> | undefined

export const overrideLicenseActionButton = (component: React.FunctionComponent<Props>): void => {
    OverrideComponent = component
}

/**
 * A button that displays the relevant action (e.g., "Upgrade" or "Start free trial") to enable a
 * feature that is not currently enabled based on the current license key (or lack thereof) or
 * product edition.
 *
 * This button is extended by the enterprise code to show more actions based on the current license
 * key.
 */
export const LicenseActionButton: React.FunctionComponent<Props> = props =>
    OverrideComponent ? (
        <OverrideComponent {...props} />
    ) : (
        <a
            href="https://docs.sourcegraph.com"
            /* TODO!(sqs) */ target="_blank"
            rel="noopener noreferrer"
            className="btn btn-sm btn-secondary"
        >
            Upgrade
        </a>
    )

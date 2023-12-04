import React from 'react'

import { Alert, Link } from '@sourcegraph/wildcard'

/**
 * Displays a warning in debug mode (which is on for local dev) that generated license keys aren't
 * actually valid. Local dev (with `sg start`) uses a test private key that does NOT correspond to
 * the license validation public key shipped with Sourcegraph instances.
 *
 * Technically it's possible to generate valid license keys in debug mode (if you manually configure
 * the right license generation private key), but it's not worth the complexity to make this alert
 * precise.
 */
export const LicenseGenerationKeyWarning: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className = '',
}) =>
    window.context?.debug ? (
        <Alert className={className} variant="warning">
            License keys generated in dev mode are <strong>NOT VALID</strong>.{' '}
            <Link to="https://sourcegraph.com/site-admin/dotcom/product/subscriptions">
                Use Sourcegraph.com to generate valid license keys.
            </Link>
        </Alert>
    ) : null

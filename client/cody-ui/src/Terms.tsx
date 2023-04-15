import React from 'react'

/**
 * A paragraph describing the Cody terms.
 */
export const Terms: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <p className={className}>
        By using Cody, you agree to its{' '}
        <a href="https://about.sourcegraph.com/terms/cody-notice">license and privacy statement</a>.
    </p>
)

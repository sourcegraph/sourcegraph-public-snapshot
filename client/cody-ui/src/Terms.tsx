import React from 'react'

import classNames from 'classnames'

export const Terms: React.FunctionComponent<{
    acceptTermsButton?: JSX.Element
    className?: string
}> = ({ acceptTermsButton, className }) => (
    <div className={classNames(className, 'non-transcript-container')}>
        <p className="terms-header-container">Notice and Usage Policies</p>
        <div className="terms-container">
            <p>
                By accepting and using Cody, you agree to the{' '}
                <a href="https://about.sourcegraph.com/terms/cody-notice">Cody Notice and Usage Policy</a>.
            </p>
        </div>
        {acceptTermsButton}
    </div>
)

import React from 'react'
import { overrideLicenseActionButton } from '../../../components/licenseActions/LicenseActionButton'

interface Props {}

/**
 * A button that displays actions for upgrading or starting a free trial for the enterprise edition.
 * It extends the OSS LicenseActionButton.
 */
const LicenseActionButton: React.FunctionComponent<Props> = () => (
    <a
        href="/subscriptions/new-trial"
        /* TODO!(sqs) */ target="_blank"
        rel="noopener noreferrer"
        className="btn btn-sm btn-secondary"
    >
        Start free trial
    </a>
)

overrideLicenseActionButton(LicenseActionButton)

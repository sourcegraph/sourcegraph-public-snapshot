import React from 'react'
import CloseIcon from 'mdi-react/CloseIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import * as GQL from '../../../../shared/src/graphql/schema'
import { serviceTypeDisplayNameAndIcon } from './GoToCodeHostAction'

interface InstallExtensionAlertProps {
    onAlertDismissed: () => void
    externalURLs: GQL.IExternalLink[]
}

export const InstallExtensionAlert: React.FunctionComponent<InstallExtensionAlertProps> = ({
    onAlertDismissed,
    externalURLs,
}) => {
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(externalURLs[0]?.serviceType)

    const Icon = icon || ExportIcon

    return (
        <div className="alert alert-info m-2 d-flex justify-content-between">
            <div className="d-flex align-items-center">
                <div className="position-relative">
                    <div className="install-extension-alert__icon-flash" />
                    <Icon className="install-extension-alert__icon" />
                </div>
                <p className="install-extension-alert__text my-0 mr-3">
                    <a href="/help/integration/browser_extension" className="alert-link">
                        Install the Sourcegraph browser extension
                    </a>{' '}
                    to add code intelligence to PRs and file views in {displayName} or any other connected code host.
                </p>
            </div>
            <button type="button" onClick={onAlertDismissed} aria-label="Close alert" className="btn btn-icon">
                <CloseIcon className="icon-inline" />
            </button>
        </div>
    )
}

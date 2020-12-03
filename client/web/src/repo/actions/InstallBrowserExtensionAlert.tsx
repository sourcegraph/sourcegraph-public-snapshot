import React from 'react'
import CloseIcon from 'mdi-react/CloseIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import * as GQL from '../../../../shared/src/graphql/schema'
import { serviceTypeDisplayNameAndIcon } from './GoToCodeHostAction'
import { eventLogger } from '../../tracking/eventLogger'
import classNames from 'classnames'

interface Props {
    onAlertDismissed: () => void
    externalURLs: GQL.IExternalLink[]
    isChrome: boolean
    codeHostIntegrationMessaging: 'browser-extension' | 'native-integration'
}

// TODO(tj): Add Firefox once the Firefox extension is back
const CHROME_EXTENSION_STORE_LINK = 'https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack'

/** Code hosts the browser extension supports */
const supportedServiceTypes = new Set<string>(['github', 'gitlab', 'phabricator', 'bitbucketServer'])

export const InstallBrowserExtensionAlert: React.FunctionComponent<Props> = ({
    onAlertDismissed,
    externalURLs,
    isChrome,
    codeHostIntegrationMessaging,
}) => {
    const externalLink = externalURLs.find(link => link.serviceType && supportedServiceTypes.has(link.serviceType))
    if (!externalLink) {
        return null
    }

    const { serviceType } = externalLink
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(serviceType)

    const Icon = icon || ExportIcon

    return (
        <div className="alert alert-info m-2 d-flex justify-content-between install-browser-extension-alert">
            <div className="d-flex align-items-center">
                <div className="position-relative">
                    <div className="install-browser-extension-alert__icon-flash" />
                    <Icon className="install-browser-extension-alert__icon" />
                </div>
                <p className="install-browser-extension-alert__text my-0 mr-3">
                    {codeHostIntegrationMessaging === 'native-integration' ? (
                        <>
                            Sourcegraph's code intelligence will follow you to your code host. Your site admin set up
                            the Sourcegraph native integration for {displayName}.{' '}
                            <a
                                className="alert-link"
                                href="https://docs.sourcegraph.com/integration/browser_extension"
                                target="_blank"
                                rel="noopener"
                            >
                                Learn more
                            </a>{' '}
                            or{' '}
                            <a className="alert-link" href={externalLink.url} target="_blank" rel="noopener">
                                try it out
                            </a>
                        </>
                    ) : isChrome ? (
                        <>
                            <a
                                href={CHROME_EXTENSION_STORE_LINK}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="alert-link"
                            >
                                Install the Sourcegraph browser extension
                            </a>{' '}
                            to add code intelligence{' '}
                            {serviceType === 'github' ||
                            serviceType === 'bitbucketServer' ||
                            serviceType === 'gitlab' ? (
                                <>to {serviceType === 'gitlab' ? 'merge requests' : 'pull requests'} and file views</>
                            ) : (
                                <>while browsing and reviewing code</>
                            )}{' '}
                            on {displayName}.
                        </>
                    ) : (
                        <>
                            Get code intelligence{' '}
                            {serviceType === 'github' ||
                            serviceType === 'bitbucketServer' ||
                            serviceType === 'gitlab' ? (
                                <>
                                    while browsing files and reviewing{' '}
                                    {serviceType === 'gitlab' ? 'merge requests' : 'pull requests'}
                                </>
                            ) : (
                                <>while browsing and reviewing code</>
                            )}{' '}
                            on {displayName}.{' '}
                            <a
                                href="/help/integration/browser_extension"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="alert-link"
                            >
                                Learn more about Sourcegraph Chrome and Firefox extensions
                            </a>
                        </>
                    )}
                </p>
            </div>
            <button
                type="button"
                onClick={onAlertDismissed}
                aria-label="Close alert"
                className="btn btn-icon test-close-alert"
            >
                <CloseIcon className="icon-inline" />
            </button>
        </div>
    )
}

// TEMPORARY: Firefox alert. Remove after the final date
interface FirefoxAlertProps {
    onAlertDismissed: () => void

    externalURLs: GQL.IExternalLink[]

    /**
     * Function that returns the current time.
     *
     * Used to assert that the alert won't show up after the final date
     */
    now?: () => Date

    /**
     * Whether a repo alert is displayed below FirefoxAddonalert
     *
     * Used to reduce distance between alerts if both are displayed at the same time
     */
    nextSiblingAlert?: boolean
}

const FIREFOX_ALERT_FINAL_DATE = new Date('December 31, 2020')

/**
 * Displays an alert to notify users that the Firefox addon is back. Doesn't do anything
 * after the final date of the campaign
 *
 * Ignore codeHostIntegrationMessaging type, this is important for all users to know
 */
export const FirefoxAddonAlert: React.FunctionComponent<FirefoxAlertProps> = ({
    now = () => new Date(),
    externalURLs,
    onAlertDismissed,
    nextSiblingAlert,
}) => {
    const currentDate = now()

    if (currentDate.getTime() > FIREFOX_ALERT_FINAL_DATE.getTime()) {
        return null
    }

    const externalLink = externalURLs.find(link => link.serviceType && supportedServiceTypes.has(link.serviceType))
    if (!externalLink) {
        return null
    }

    const { serviceType } = externalLink
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(serviceType)

    const Icon = icon || ExportIcon

    return (
        <div
            className={classNames(
                'alert alert-info d-flex justify-content-between install-browser-extension-alert',
                nextSiblingAlert ? 'mx-2 mt-2 mb-0' : 'm-2'
            )}
        >
            <div className="d-flex align-items-center">
                <div className="position-relative">
                    <div className="install-browser-extension-alert__icon-flash" />
                    <Icon className="install-browser-extension-alert__icon" />
                </div>
                <p className="install-browser-extension-alert__text my-0 mr-3">
                    <strong>Sourcegraph Firefox add-on is back!</strong> 🎉️ To add code intelligence{' '}
                    {serviceType === 'github' || serviceType === 'bitbucketServer' || serviceType === 'gitlab' ? (
                        <>to {serviceType === 'gitlab' ? 'merge requests' : 'pull requests'} and file views</>
                    ) : (
                        <>while browsing and reviewing code</>
                    )}{' '}
                    on {displayName} or any other connected code host,{' '}
                    <a
                        href="/help/integration/browser_extension"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="alert-link"
                        onClick={onInstallLinkClick}
                    >
                        install the add-on
                    </a>
                </p>
            </div>
            <button
                type="button"
                onClick={onAlertDismissed}
                aria-label="Close alert"
                className="btn btn-icon test-close-alert"
            >
                <CloseIcon className="icon-inline" />
            </button>
        </div>
    )
}

const onInstallLinkClick = (): void => {
    eventLogger.log('FirefoxAlertInstallClicked')
}

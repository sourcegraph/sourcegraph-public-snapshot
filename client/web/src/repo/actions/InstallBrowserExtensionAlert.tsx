import React from 'react'
import CloseIcon from 'mdi-react/CloseIcon'
import ExportIcon from 'mdi-react/ExportIcon'
import * as GQL from '../../../../shared/src/graphql/schema'
import { serviceTypeDisplayNameAndIcon } from './GoToCodeHostAction'
import { eventLogger } from '../../tracking/eventLogger'
interface InstallBrowserExtensionAlertProps {
    onAlertDismissed: () => void
    externalURLs: GQL.IExternalLink[]
    isChrome: boolean
    codeHostIntegrationMessaging: 'browser-extension' | 'native-integration'

    // TEMPORARY
    showFirefoxAddonAlert?: boolean
}

// TODO(tj): Add Firefox once the Firefox extension is listed on AMO again
const CHROME_EXTENSION_STORE_LINK = 'https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack'

/** Code hosts the browser extension supports */
const supportedServiceTypes = new Set<string>(['github', 'gitlab', 'phabricator', 'bitbucketServer'])

export const InstallBrowserExtensionAlert: React.FunctionComponent<InstallBrowserExtensionAlertProps> = ({
    onAlertDismissed,
    externalURLs,
    isChrome,
    codeHostIntegrationMessaging,
    showFirefoxAddonAlert,
}) => {
    const externalLink = externalURLs.find(link => link.serviceType && supportedServiceTypes.has(link.serviceType))
    if (!externalLink) {
        return null
    }

    const { serviceType } = externalLink
    const { displayName, icon } = serviceTypeDisplayNameAndIcon(serviceType)

    const Icon = icon || ExportIcon

    const renderedIcon = <Icon className="install-browser-extension-alert__icon" />

    if (showFirefoxAddonAlert) {
        return (
            <FirefoxAddonAlert
                onAlertDismissed={onAlertDismissed}
                serviceType={serviceType}
                displayName={displayName}
                icon={renderedIcon}
            />
        )
    }

    return (
        <div className="alert alert-info m-2 d-flex justify-content-between install-browser-extension-alert">
            <div className="d-flex align-items-center">
                <div className="position-relative">
                    <div className="install-browser-extension-alert__icon-flash" />
                    {renderedIcon}
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

    displayName: string

    serviceType: string | null

    icon: JSX.Element
}

const FIREFOX_ALERT_START_DATE = new Date('December 6, 2020')
export const FIREFOX_ALERT_FINAL_DATE = new Date('December 31, 2020')

export function isFirefoxCampaignActive(currentMs: number): boolean {
    return currentMs < FIREFOX_ALERT_FINAL_DATE.getTime() && currentMs > FIREFOX_ALERT_START_DATE.getTime()
}

/**
 * Ignore codeHostIntegrationMessaging type, this is important for all users to know
 */
export const FirefoxAddonAlert: React.FunctionComponent<FirefoxAlertProps> = ({
    onAlertDismissed,
    displayName,
    serviceType,
    icon,
}) => (
    <div className="alert alert-info m-2 d-flex justify-content-between install-browser-extension-alert">
        <div className="d-flex align-items-center">
            <div className="position-relative">
                <div className="install-browser-extension-alert__icon-flash" />
                {icon}
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

const onInstallLinkClick = (): void => {
    eventLogger.log('FirefoxAlertInstallClicked')
}

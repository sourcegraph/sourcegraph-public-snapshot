import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

import { ExternalLinkFields, ExternalServiceKind } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { serviceKindDisplayNameAndIcon } from './GoToCodeHostAction'

interface Props {
    onAlertDismissed: () => void
    externalURLs: ExternalLinkFields[]
    isChrome: boolean
    codeHostIntegrationMessaging: 'browser-extension' | 'native-integration'
    showFirefoxAddonAlert?: boolean
}

// TODO(tj): Add Firefox once the Firefox extension is back
const CHROME_EXTENSION_STORE_LINK = 'https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack'

/** Code hosts the browser extension supports */
const supportedServiceTypes = new Set<string>([
    ExternalServiceKind.GITHUB,
    ExternalServiceKind.GITLAB,
    ExternalServiceKind.PHABRICATOR,
    ExternalServiceKind.BITBUCKETSERVER,
])

export const InstallBrowserExtensionAlert: React.FunctionComponent<Props> = ({
    onAlertDismissed,
    externalURLs,
    isChrome,
    codeHostIntegrationMessaging,
    showFirefoxAddonAlert,
}) => {
    const externalLink = externalURLs.find(link => link.serviceKind && supportedServiceTypes.has(link.serviceKind))
    if (!externalLink) {
        return null
    }

    const { serviceKind } = externalLink
    const { displayName } = serviceKindDisplayNameAndIcon(serviceKind)

    if (showFirefoxAddonAlert) {
        return <FirefoxAddonAlert onAlertDismissed={onAlertDismissed} displayName={displayName} />
    }

    return (
        <div
            className="alert alert-info m-2 d-flex justify-content-between flex-shrink-0"
            data-testid="install-browser-extension-alert"
        >
            <div className="d-flex align-items-center">
                <p className="my-0 mr-3">
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
                            {serviceKind === ExternalServiceKind.GITHUB ||
                            serviceKind === ExternalServiceKind.BITBUCKETSERVER ||
                            serviceKind === ExternalServiceKind.GITLAB ? (
                                <>
                                    to {serviceKind === ExternalServiceKind.GITLAB ? 'merge requests' : 'pull requests'}{' '}
                                    and file views
                                </>
                            ) : (
                                <>while browsing and reviewing code</>
                            )}{' '}
                            on {displayName}.
                        </>
                    ) : (
                        <>
                            Get code intelligence{' '}
                            {serviceKind === ExternalServiceKind.GITHUB ||
                            serviceKind === ExternalServiceKind.BITBUCKETSERVER ||
                            serviceKind === ExternalServiceKind.GITLAB ? (
                                <>
                                    while browsing files and reviewing{' '}
                                    {serviceKind === ExternalServiceKind.GITLAB ? 'merge requests' : 'pull requests'}
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

interface FirefoxAlertProps {
    onAlertDismissed: () => void
    displayName: string
}

const FIREFOX_ALERT_START_DATE = new Date('July 16, 2021')
export const FIREFOX_ALERT_FINAL_DATE = new Date('October 18, 2021')

export function isFirefoxCampaignActive(currentMs: number): boolean {
    return currentMs < FIREFOX_ALERT_FINAL_DATE.getTime() && currentMs > FIREFOX_ALERT_START_DATE.getTime()
}

export const FirefoxAddonAlert: React.FunctionComponent<FirefoxAlertProps> = ({ onAlertDismissed, displayName }) => (
    <div className="alert alert-info m-2 d-flex justify-content-between flex-shrink-0 percy-hide">
        <div>
            <p className="font-weight-medium my-0 mr-3">
                Sourcegraph is back at{' '}
                <a
                    href="https://addons.mozilla.org/en-US/firefox/addon/sourcegraph-for-firefox"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="alert-link"
                    onClick={onInstallLinkClick}
                >
                    Firefox Add-ons
                </a>{' '}
                🎉️
            </p>
            <p className="mt-1 mb-0">
                If you already have the local version,{' '}
                <a
                    href="https://docs.sourcegraph.com/integration/migrating_firefox_extension"
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={onInstallLinkClick}
                >
                    make sure to upgrade
                </a>
                . The extension adds code intelligence to code views on {displayName} or any other connected code host.
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

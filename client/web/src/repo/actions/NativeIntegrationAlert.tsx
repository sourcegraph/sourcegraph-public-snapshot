import React, { useEffect } from 'react'

import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'
import { AlertLink } from '@sourcegraph/wildcard'

import { ExtensionRadialGradientIcon } from '../../components/CtaIcons'
import { ExternalLinkFields, ExternalServiceKind } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { serviceKindDisplayNameAndIcon } from './GoToCodeHostAction'

export interface NativeIntegrationAlertProps {
    className?: string
    onAlertDismissed: () => void
    externalURLs: ExternalLinkFields[]
}

/** Code hosts the browser extension supports */
const supportedServiceTypes = new Set<string>([
    ExternalServiceKind.GITHUB,
    ExternalServiceKind.GITLAB,
    ExternalServiceKind.PHABRICATOR,
    ExternalServiceKind.BITBUCKETSERVER,
])

export const NativeIntegrationAlert: React.FunctionComponent<NativeIntegrationAlertProps> = ({
    className,
    onAlertDismissed,
    externalURLs,
}) => {
    useEffect(() => {
        eventLogger.log('NativeIntegrationInstallShown')
    }, [])

    const externalLink = externalURLs.find(link => link.serviceKind && supportedServiceTypes.has(link.serviceKind))
    if (!externalLink) {
        return null
    }

    const { serviceKind } = externalLink
    const { displayName } = serviceKindDisplayNameAndIcon(serviceKind)

    return (
        <CtaAlert
            title={`Your site admin set up the Sourcegraph native integration for ${displayName}.`}
            description={
                <>
                    Sourcegraph's code intelligence will follow you to your code host.{' '}
                    <AlertLink
                        to="https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=inproduct-cta&utm_medium=direct_traffic&utm_source=search-results-cta&utm_term=null&utm_content=install-browser-exten"
                        target="_blank"
                        rel="noopener"
                    >
                        Learn more
                    </AlertLink>
                </>
            }
            cta={{ label: 'Try it out', href: externalLink.url, onClick: installLinkClickHandler }}
            icon={<ExtensionRadialGradientIcon />}
            className={className}
            onClose={onAlertDismissed}
        />
    )
}

const installLinkClickHandler = (): void => {
    eventLogger.log('NativeIntegrationInstallClicked')
}

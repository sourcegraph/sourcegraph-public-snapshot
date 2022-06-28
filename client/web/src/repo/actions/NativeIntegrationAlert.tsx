import React, { useEffect, useCallback, useMemo } from 'react'

import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'
import { AlertLink } from '@sourcegraph/wildcard'

import { ExtensionRadialGradientIcon } from '../../components/CtaIcons'
import { ExternalLinkFields, ExternalServiceKind } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { serviceKindDisplayNameAndIcon } from './GoToCodeHostAction'

export interface NativeIntegrationAlertProps {
    className?: string
    page: 'search' | 'file'
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

export const NativeIntegrationAlert: React.FunctionComponent<React.PropsWithChildren<NativeIntegrationAlertProps>> = ({
    className,
    page,
    onAlertDismissed,
    externalURLs,
}) => {
    const args = useMemo(() => ({ page }), [page])
    useEffect(() => {
        eventLogger.log('NativeIntegrationInstallShown', args, args)
    }, [args])

    const installLinkClickHandler = useCallback((): void => {
        eventLogger.log('NativeIntegrationInstallClicked', args, args)
    }, [args])

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
                        to="https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=search-results-cta&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=install-browser-exten"
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

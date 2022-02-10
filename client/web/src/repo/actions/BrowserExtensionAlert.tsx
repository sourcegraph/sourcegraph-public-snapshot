import React, { useEffect } from 'react'

import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'

import { ExtensionRadialGradientIcon } from '../../components/CtaIcons'
import { eventLogger } from '../../tracking/eventLogger'

interface Props {
    className?: string
    page: 'search' | 'file'
    onAlertDismissed: () => void
}

export const BrowserExtensionAlert: React.FunctionComponent<Props> = ({ className, page, onAlertDismissed }) => {
    useEffect(() => {
        eventLogger.log('InstallBrowserExtensionCTAShown', undefined, { page })
    }, [page])

    return (
        <CtaAlert
            title="Install the Sourcegraph browser extension"
            description="Add code intelligence to pull requests and file views on GitHub, GitLab, Bitbucket Server, and Phabricator"
            cta={{
                label: 'Learn more about the extension',
                href:
                    'https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=inproduct-cta&utm_medium=direct_traffic&utm_source=search-results-cta&utm_term=null&utm_content=install-browser-exten',
                onClick: () => onBrowserExtensionClick(page),
            }}
            icon={<ExtensionRadialGradientIcon />}
            className={className}
            onClose={onAlertDismissed}
        />
    )
}

const onBrowserExtensionClick = (page: 'search' | 'file'): void => {
    eventLogger.log('InstallBrowserExtensionCTAClicked', undefined, { page })
}

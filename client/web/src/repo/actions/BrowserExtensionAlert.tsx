import React, { useCallback, useEffect, useMemo } from 'react'

import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'

import { ExtensionRadialGradientIcon } from '../../components/CtaIcons'
import { IS_CHROME } from '../../marketing/util'
import { eventLogger } from '../../tracking/eventLogger'

interface Props {
    className?: string
    page: 'search' | 'file'
    onAlertDismissed: () => void
}

const CHROME_LINK = 'https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack'
const SAFARI_LINK = 'https://apps.apple.com/us/app/sourcegraph-for-safari/id1543262193'
const FIREFOX_LINK = 'https://addons.mozilla.org/en-US/firefox/addon/sourcegraph-for-firefox/'

export const BrowserExtensionAlert: React.FunctionComponent<Props> = ({ className, page, onAlertDismissed }) => {
    const args = useMemo(() => ({ page }), [page])

    useEffect(() => {
        eventLogger.log('InstallBrowserExtensionCTAShown', args, args)
    }, [args])

    const onBrowserExtensionPrimaryClick = useCallback((): void => {
        eventLogger.log('InstallBrowserExtensionCTAClicked', args, args)
    }, [args])

    const onBrowserExtensionSecondaryClick = useCallback((): void => {
        eventLogger.log('InstallBrowserExtensionLearnClicked', args, args)
    }, [args])

    return (
        <CtaAlert
            title="Get more from Sourcegraph with the browser extension"
            description="Add code intelligence to pull requests and file views on GitHub, GitLab, Bitbucket Server, and Phabricator."
            cta={{
                label: 'Install now',
                href: IS_CHROME
                    ? SAFARI_LINK
                    : 'https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=search-results-cta&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=install-browser-exten',
                onClick: onBrowserExtensionPrimaryClick,
            }}
            secondary={{
                label: 'Learn more',
                href:
                    'https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=search-results-cta&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=install-browser-exten',
                onClick: onBrowserExtensionSecondaryClick,
            }}
            icon={<ExtensionRadialGradientIcon />}
            className={className}
            onClose={onAlertDismissed}
        />
    )
}

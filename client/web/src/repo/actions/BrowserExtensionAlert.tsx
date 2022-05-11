import React, { useCallback, useEffect, useMemo } from 'react'

import { getBrowserName } from '@sourcegraph/common'
import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'

import { ExtensionRadialGradientIcon } from '../../components/CtaIcons'
import { eventLogger } from '../../tracking/eventLogger'

interface Props {
    className?: string
    page: 'search' | 'file'
    onAlertDismissed: () => void
}

const LEARN_MORE_URL =
    'https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=search-results-cta&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=install-browser-exten'

const BROWSER_NAME = getBrowserName()
const BROWSER_NAME_TO_URL = {
    chrome: 'https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack',
    firefox: 'https://addons.mozilla.org/en-US/firefox/addon/sourcegraph-for-firefox/',
    safari: 'https://apps.apple.com/us/app/sourcegraph-for-safari/id1543262193',
    other: LEARN_MORE_URL,
}

export const BrowserExtensionAlert: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    page,
    onAlertDismissed,
}) => {
    const args = useMemo(() => ({ page, browser: BROWSER_NAME }), [page])

    useEffect(() => {
        eventLogger.log('InstallBrowserExtensionCTAShown', args, args)
    }, [args])

    const onBrowserExtensionPrimaryClick = useCallback((): void => {
        eventLogger.log('InstallBrowserExtensionCTAClicked', args, args)
    }, [args])

    const onBrowserExtensionSecondaryClick = useCallback((): void => {
        eventLogger.log('InstallBrowserExtensionLearnClicked', args, args)
    }, [args])

    const cta = {
        label: BROWSER_NAME !== 'other' ? 'Install now' : 'Learn more',
        href: BROWSER_NAME_TO_URL[BROWSER_NAME],
        onClick: BROWSER_NAME !== 'other' ? onBrowserExtensionPrimaryClick : onBrowserExtensionSecondaryClick,
    }

    const secondary =
        BROWSER_NAME !== 'other'
            ? {
                  label: 'Learn more',
                  href: LEARN_MORE_URL,
                  onClick: onBrowserExtensionSecondaryClick,
              }
            : undefined

    return (
        <CtaAlert
            title="Get more from Sourcegraph with the browser extension"
            description="Add code intelligence to pull requests and file views on GitHub, GitLab, Bitbucket Server, and Phabricator."
            cta={cta}
            secondary={secondary}
            icon={<ExtensionRadialGradientIcon />}
            className={className}
            onClose={onAlertDismissed}
        />
    )
}

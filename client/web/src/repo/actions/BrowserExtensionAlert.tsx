import React, { useCallback, useEffect, useMemo } from 'react'

import { CtaAlert } from '@sourcegraph/shared/src/components/CtaAlert'

import { ExtensionRadialGradientIcon } from '../../components/CtaIcons'
import { eventLogger } from '../../tracking/eventLogger'

interface Props {
    className?: string
    page: 'search' | 'file'
    onAlertDismissed: () => void
}

const UA = navigator.userAgent
const BROWSER: 'chrome' | 'safari' | 'firefox' | 'other' = UA.match(/chrome|chromium|crios/i)
    ? 'chrome'
    : UA.match(/firefox|fxios/i)
    ? 'firefox'
    : UA.match(/safari/i)
    ? 'safari'
    : 'other'

const CHROME_LINK = 'https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack'
const SAFARI_LINK = 'https://apps.apple.com/us/app/sourcegraph-for-safari/id1543262193'
const FIREFOX_LINK = 'https://addons.mozilla.org/en-US/firefox/addon/sourcegraph-for-firefox/'

const LEARN_MORE_LINK =
    'https://docs.sourcegraph.com/integration/browser_extension?utm_campaign=search-results-cta&utm_medium=direct_traffic&utm_source=in-product&utm_term=null&utm_content=install-browser-exten'

export const BrowserExtensionAlert: React.FunctionComponent<Props> = ({ className, page, onAlertDismissed }) => {
    const args = useMemo(() => ({ page, browser: BROWSER }), [page])

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
        label: BROWSER !== 'other' ? 'Install now' : 'Learn more',
        href:
            BROWSER === 'chrome'
                ? CHROME_LINK
                : BROWSER === 'safari'
                ? SAFARI_LINK
                : BROWSER === 'firefox'
                ? FIREFOX_LINK
                : LEARN_MORE_LINK,
        onClick: BROWSER !== 'other' ? onBrowserExtensionPrimaryClick : onBrowserExtensionSecondaryClick,
    }

    const secondary =
        BROWSER !== 'other'
            ? {
                  label: 'Learn more',
                  href: LEARN_MORE_LINK,
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

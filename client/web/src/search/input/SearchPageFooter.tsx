import classNames from 'classnames'
import React, { useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { eventLogger } from '../../tracking/eventLogger'

export const SearchPageFooter: React.FunctionComponent<{ className?: string }> = ({ className }) => {
    const logLinkClicked = useCallback(
        (name: string) => () => {
            // TODO(farhan): these events are named with the name of the CTA because
            // we currently can't collect arguments in BQ. Remove the name of CTA
            // once we can collect arguments.
            eventLogger.log('HomepageFooterCTASelected', { name }, { name })
        },
        []
    )

    return (
        <footer className={classNames(className, 'd-flex flex-column flex-lg-row align-items-center')}>
            <h4 className="mb-2 mb-lg-0">Explore and extend</h4>
            <span className="d-flex flex-column flex-md-row align-items-center">
                <span className="d-flex flex-row mb-2 mb-md-0">
                    <Link
                        className="px-3"
                        to="https://docs.sourcegraph.com/integration/browser_extension"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={logLinkClicked('BrowserExtensions')}
                    >
                        Browser extensions
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                    <Link
                        className="px-3"
                        to="/extensions"
                        target="_blank"
                        onClick={logLinkClicked('SourcegraphExtensions')}
                    >
                        Sourcegraph extensions
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                </span>
                <span className="d-flex flex-row">
                    <Link
                        className="px-3"
                        to="https://docs.sourcegraph.com/integration/editor"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={logLinkClicked('EditorPlugins')}
                    >
                        Editor plugins
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                    <Link
                        className="pl-3"
                        to="https://docs.sourcegraph.com/admin/external_service"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={logLinkClicked('CodeHostIntegrations')}
                    >
                        Code host integrations
                    </Link>
                </span>
            </span>
        </footer>
    )
}

import classNames from 'classnames'
import React from 'react'
import { Link } from '../../../../shared/src/components/Link'

export const SearchPageFooter: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <footer className={classNames(className, 'd-flex flex-row')}>
        <h4 className="mb-0">Explore and Extend:</h4>
        <Link
            className="border-right px-3"
            to="https://docs.sourcegraph.com/integration/browser_extension"
            rel="noopener noreferrer"
            target="_blank"
        >
            Browser extensions
        </Link>
        <Link className="border-right px-3" to="/extensions" target="_blank">
            Sourcegraph extensions
        </Link>
        <Link
            className="border-right px-3"
            to="https://docs.sourcegraph.com/integration/editor"
            rel="noopener noreferrer"
            target="_blank"
        >
            Editor plugins
        </Link>
        <Link
            className="pl-3"
            to="https://docs.sourcegraph.com/admin/external_service"
            rel="noopener noreferrer"
            target="_blank"
        >
            Code host integrations
        </Link>
    </footer>
)

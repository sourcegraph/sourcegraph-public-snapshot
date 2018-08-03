import * as React from 'react'
import { Link } from 'react-router-dom'

/** The common empty state for extensions with a CTA to the extension registry. */
export const ExtensionsEmptyState: React.SFC<{
    className?: string
    onClick?: () => void
}> = ({ className = 'px-3 py-4', onClick }) => (
    <div className={`${className} text-center bg-striped-secondary`}>
        <h4 className="text-muted mb-3">Extensions add new features to Sourcegraph.</h4>
        <Link to="/extensions" className="btn btn-primary" onClick={onClick}>
            Add extensions from registry
        </Link>
    </div>
)

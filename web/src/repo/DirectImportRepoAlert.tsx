import InfoIcon from '@sourcegraph/icons/lib/Info'
import * as React from 'react'

export const DirectImportRepoAlert: React.StatelessComponent<{ className?: string }> = ({ className = '' }) => (
    <>
        {!window.context.isRunningDataCenter && (
            <div className={`alert alert-info ${className}`}>
                <InfoIcon className="icon-inline" /> Very large repository? See{' '}
                <a href="https://about.sourcegraph.com/docs/config/repositories#add-repositories-already-cloned-to-disk">
                    how to reuse an existing local clone
                </a>{' '}
                to speed this up.
            </div>
        )}
    </>
)

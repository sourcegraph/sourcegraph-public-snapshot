import InfoIcon from '@sourcegraph/icons/lib/Info'
import * as React from 'react'

export const DirectImportRepoAlert: React.StatelessComponent<{ className?: string }> = ({ className = '' }) => (
    <>
        {!window.context.isRunningDataCenter && (
            <div className={`alert alert-info ${className}`}>
                <InfoIcon className="icon-inline" /> Admins can directly import large repos that are already on the host
                machine by following{' '}
                <a href="https://about.sourcegraph.com/docs/config/repositories#add-repositories-already-cloned-to-disk">
                    these instructions
                </a>.
            </div>
        )}
    </>
)

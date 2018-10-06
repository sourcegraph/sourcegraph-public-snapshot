import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import * as React from 'react'

export const DirectImportRepoAlert: React.StatelessComponent<{ className?: string }> = ({ className = '' }) => (
    <>
        {!window.context.isClusterDeployment && (
            <div className={`alert alert-info ${className}`}>
                <InformationOutlineIcon className="icon-inline" /> Very large repository? See{' '}
                <a href="https://about.sourcegraph.com/docs/config/repositories#add-repositories-already-cloned-to-disk">
                    how to reuse an existing local clone
                </a>{' '}
                to speed this up.
            </div>
        )}
    </>
)

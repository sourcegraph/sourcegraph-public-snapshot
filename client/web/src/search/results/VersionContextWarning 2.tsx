import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

export const VersionContextWarning: React.FunctionComponent<{
    versionContext?: string
    onDismissWarning: () => void
}> = ({ versionContext, onDismissWarning }) => (
    <div className="mt-2 mx-2">
        <div className="d-flex alert alert-warning mb-0 justify-content-between align-items-center">
            <div>
                This link changed your version context to <strong>{versionContext || 'default'}</strong>. You can switch
                contexts with the selector to the left of the search bar.
            </div>
            <button type="button" className="btn p-0" onClick={onDismissWarning}>
                <CloseIcon className="icon-inline ml-2" />
            </button>
        </div>
    </div>
)

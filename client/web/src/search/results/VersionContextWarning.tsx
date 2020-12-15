import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

export const VersionContextWarning: React.FunctionComponent<{
    versionContext?: string
    onDismissWarning: () => void
}> = ({ versionContext, onDismissWarning }) => (
    <div className="mt-2 mx-2">
        <div className="d-flex alert alert-warning mb-0 justify-content-between">
            <div>
                <AlertOutlineIcon className="icon-inline mr-2" />
                This link changed your version context to <strong>{versionContext || 'default'}</strong>. You can switch
                contexts with the selector to the left of the search bar.
            </div>
            <div onClick={onDismissWarning}>
                <CloseIcon className="icon-inline ml-2" />
            </div>
        </div>
    </div>
)

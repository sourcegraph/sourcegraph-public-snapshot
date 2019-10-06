import React from 'react'
import AlertIcon from 'mdi-react/AlertIcon'

interface Props {
    noun: string
    action: JSX.Element
}

export const IsPendingExternalCreationTimelineBox: React.FunctionComponent<Props> = ({ noun, action }) => (
    <div className="d-flex align-items-center bg-body border mt-5 p-4">
        <AlertIcon className="icon-inline h4 mb-0 mr-3" />
        <h3 className="flex-1 mb-0">This {noun} preview has not yet been created on the external service</h3>
        {action}
    </div>
)

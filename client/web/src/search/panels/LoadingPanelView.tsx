import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export const LoadingPanelView: React.FunctionComponent<{ text: string }> = ({ text }) => (
    <div className="d-flex justify-content-center align-items-center panel-container__empty-container panel-container__loading-container">
        <LoadingSpinner className="icon-inline" />
        <span className="text-muted">{text}</span>
    </div>
)

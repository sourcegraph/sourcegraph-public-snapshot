import * as React from 'react'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export const LoadingModal: React.FunctionComponent<{ text: string }> = ({ text }) => (
    <div className="d-flex justify-content-center align-items-center panel-container__empty-container">
        <div className="icon-inline">
            <LoadingSpinner />
        </div>
        <span className="text-muted">{text}</span>
    </div>
)

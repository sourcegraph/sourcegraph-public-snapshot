import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

export const ExtensionsLoadingPanelView: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <div className="panel">
        <div className={`panel__empty ${className}`}>
            <LoadingSpinner />
            <span className="mx-2">Loading Sourcegraph extensions</span>
            <PuzzleIcon className="icon-inline" />
        </div>
    </div>
)

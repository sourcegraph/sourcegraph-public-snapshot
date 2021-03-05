import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'
import { Link } from '../../../../../shared/src/components/Link'

export const ExtensionsLoadingPanelView: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <div className={`panel__empty ${className}`}>
        <LoadingSpinner />
        <span className="mx-2">
            Loading <Link to="/extensions">Sourcegraph extensions</Link>
        </span>
        <PuzzleIcon className="icon-inline" />
    </div>
)

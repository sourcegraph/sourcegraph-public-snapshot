import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { EmptyPanelView } from './EmptyPanelView'

interface ExtensionsLoadingPanelViewProps {
    className?: string
}

export const ExtensionsLoadingPanelView: React.FunctionComponent<ExtensionsLoadingPanelViewProps> = props => {
    const { className } = props

    return (
        <EmptyPanelView className={className}>
            <LoadingSpinner />
            <span className="mx-2">Loading Sourcegraph extensions</span>
            <PuzzleIcon className="icon-inline" />
        </EmptyPanelView>
    )
}

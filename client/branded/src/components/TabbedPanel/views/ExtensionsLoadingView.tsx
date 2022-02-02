import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { EmptyPanelView } from './EmptyPanelView'

interface ExtensionsLoadingPanelViewProps {
    className?: string
}

export const ExtensionsLoadingPanelView: React.FunctionComponent<ExtensionsLoadingPanelViewProps> = props => {
    const { className } = props

    return (
        <EmptyPanelView className={className}>
            <LoadingSpinner inline={false} />
            <span className="mx-2">Loading Sourcegraph extensions</span>
            <PuzzleIcon className="icon-inline" />
        </EmptyPanelView>
    )
}

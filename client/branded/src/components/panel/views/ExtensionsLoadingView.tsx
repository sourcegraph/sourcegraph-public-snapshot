import React from 'react'

import PuzzleIcon from 'mdi-react/PuzzleIcon'

import { LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { EmptyPanelView } from './EmptyPanelView'

interface ExtensionsLoadingPanelViewProps {
    className?: string
}

export const ExtensionsLoadingPanelView: React.FunctionComponent<
    React.PropsWithChildren<ExtensionsLoadingPanelViewProps>
> = props => {
    const { className } = props

    return (
        <EmptyPanelView className={className}>
            <LoadingSpinner inline={false} />
            <span className="mx-2">Loading Sourcegraph extensions</span>
            <Icon role="img" as={PuzzleIcon} aria-hidden={true} />
        </EmptyPanelView>
    )
}

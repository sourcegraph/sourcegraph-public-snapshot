import { Icons, IconButton } from '@storybook/components'
import React, { ReactElement } from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

export const RedesignToggleStorybook = (): ReactElement => {
    const [isRedesignEnabled, setIsRedesignEnabled] = useRedesignToggle()

    const handleRedesignToggle = (): void => {
        setIsRedesignEnabled(!isRedesignEnabled)
    }

    return (
        <IconButton
            key="redesign-toolbar"
            active={isRedesignEnabled}
            title={isRedesignEnabled ? 'Disable redesign theme' : 'Enable redesign theme'}
            // eslint-disable-next-line react/jsx-no-bind
            onClick={handleRedesignToggle}
        >
            <Icons icon="beaker" />
        </IconButton>
    )
}

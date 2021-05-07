import React from 'react'

import { useRedesignToggle } from '../util/useRedesignToggle'

interface RedesignToggleProviderProps {
    children(renderProps: {
        isRedesignEnabled: boolean
        setIsRedesignEnabled(isRedesignEnabled: boolean): void
    }): React.ReactElement
}

export const RedesignToggleProvider: React.FunctionComponent<RedesignToggleProviderProps> = ({ children }) => {
    const [isRedesignEnabled, setIsRedesignEnabled] = useRedesignToggle()

    return children({ isRedesignEnabled, setIsRedesignEnabled })
}

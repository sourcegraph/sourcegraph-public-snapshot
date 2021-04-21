import React, { useCallback } from 'react'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { useRedesignToggle, REDESIGN_CLASS_NAME } from '@sourcegraph/shared/src/util/useRedesignToggle'

export const RedesignToggle: React.FunctionComponent = () => {
    const [isRedesignEnabled, setIsRedesignEnabled] = useRedesignToggle()

    const handleRedesignToggle = useCallback((): void => {
        setIsRedesignEnabled(!isRedesignEnabled)
        document.documentElement.classList.toggle(REDESIGN_CLASS_NAME, !isRedesignEnabled)
    }, [isRedesignEnabled, setIsRedesignEnabled])

    return (
        <div className="px-2 py-1">
            <div className="d-flex align-items-center">
                <div className="mr-2">Redesign enabled</div>
                <Toggle title="Redesign theme enabled" value={isRedesignEnabled} onToggle={handleRedesignToggle} />
            </div>
        </div>
    )
}

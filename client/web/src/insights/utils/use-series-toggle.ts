import { type Dispatch, type SetStateAction, useCallback, useState } from 'react'

export interface UseSeriesToggleReturn {
    /* List of currently selected series ids */
    selectedSeriesIds: string[]
    hoveredId: string | undefined

    // These functions are exposed to keep all of the state
    // contained in the hook
    isSeriesSelected: (id: string) => boolean
    isSeriesHovered: (id: string) => boolean
    setHoveredId: Dispatch<SetStateAction<string | undefined>>
    toggle: (id: string, availableSeriesIds: string[]) => void
    setSelectedSeriesIds: Dispatch<SetStateAction<string[]>>
}

// Used when clicking legend items in a line chart. This hook manages the currently
// selected data series as well as the currently hovered legend item.
/**
 * @returns helper tools for managing the currently selected and hovered series
 */
export const useSeriesToggle = (): UseSeriesToggleReturn => {
    const [selectedSeriesIds, setSelectedSeriesIds] = useState<string[]>([])
    const [hoveredId, setHoveredId] = useState<string | undefined>()

    const selectSeries = (seriesId: string, availableSeriesIds: string[]): void => {
        const nextSelectedSeriesIds = [...selectedSeriesIds, seriesId]

        // Reset the selected series if the user is about to select all of them
        if (nextSelectedSeriesIds.length === availableSeriesIds.length) {
            return setSelectedSeriesIds([])
        }

        setSelectedSeriesIds(nextSelectedSeriesIds)
    }

    const deselectSeries = (seriesId: string): void =>
        setSelectedSeriesIds(selectedSeriesIds.filter(id => id !== seriesId))

    const toggle = (seriesId: string, availableSeriesIds: string[]): void => {
        if (selectedSeriesIds.includes(seriesId)) {
            return deselectSeries(seriesId)
        }

        selectSeries(seriesId, availableSeriesIds)
    }

    const isSelected = useCallback(
        (seriesId: string): boolean => {
            // Return true for all series if no series are selected
            // This is because we only want to hide series if something is
            // specifically selected. Otherwise, they should all be "highlighted"
            if (selectedSeriesIds.length === 0) {
                return true
            }

            return selectedSeriesIds.includes(seriesId)
        },
        [selectedSeriesIds]
    )

    return {
        // state
        selectedSeriesIds,
        hoveredId,

        // functions
        isSeriesSelected: isSelected,
        isSeriesHovered: useCallback((seriesId: string) => seriesId === hoveredId, [hoveredId]),
        setHoveredId,
        toggle,
        setSelectedSeriesIds,
    }
}

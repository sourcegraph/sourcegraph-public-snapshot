import { Dispatch, SetStateAction, useState } from 'react'

interface UseSeriesToggleReturn {
    /* List of currently selected series ids */
    selectedSeriesIds: string[]

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

    const selectSeries = (seriesId: string): void => setSelectedSeriesIds([...selectedSeriesIds, seriesId])
    const deselectSeries = (seriesId: string): void =>
        setSelectedSeriesIds(selectedSeriesIds.filter(id => id !== seriesId))
    const toggle = (seriesId: string, availableSeriesIds: string[]): void => {
        // Reset the selected series if the user is about to select all of them
        if (selectedSeriesIds.length === availableSeriesIds.length - 1) {
            return setSelectedSeriesIds([])
        }
        return selectedSeriesIds.includes(seriesId) ? deselectSeries(seriesId) : selectSeries(seriesId)
    }
    const isSelected = (seriesId: string): boolean => {
        // Return true for all series if no series are selected
        // This is because we only want to hide series if something is
        // specifically selected. Otherwise they should all be "highlighted"
        if (selectedSeriesIds.length === 0) {
            return true
        }

        return selectedSeriesIds.includes(seriesId)
    }

    return {
        // state
        selectedSeriesIds,

        // functions
        isSeriesSelected: isSelected,
        isSeriesHovered: (seriesId: string) => seriesId === hoveredId,
        setHoveredId,
        toggle,
        setSelectedSeriesIds,
    }
}

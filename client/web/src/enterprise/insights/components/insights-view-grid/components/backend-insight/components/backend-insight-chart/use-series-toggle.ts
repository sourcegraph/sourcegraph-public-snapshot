import { useState } from 'react'

type SeriesId = string | number
type SeriesIds = SeriesId[]

interface UseSeriesToggleReturn {
    toggle: (id: SeriesId) => void
    selectedSeriesIds: SeriesIds
}

export const useSeriesToggle = (currentSelectedSeriesIds: SeriesIds): UseSeriesToggleReturn => {
    const [selectedSeriesIds, setSelectedSeriesIds] = useState<SeriesIds>(currentSelectedSeriesIds)

    const selectSeries = (seriesId: SeriesId): void => setSelectedSeriesIds([...selectedSeriesIds, seriesId])
    const deselectSeries = (seriesId: SeriesId): void =>
        setSelectedSeriesIds(selectedSeriesIds.filter(id => id !== seriesId))
    const toggle = (seriesId: SeriesId): void =>
        selectedSeriesIds.includes(seriesId) ? deselectSeries(seriesId) : selectSeries(seriesId)

    return {
        toggle,
        selectedSeriesIds,
    }
}

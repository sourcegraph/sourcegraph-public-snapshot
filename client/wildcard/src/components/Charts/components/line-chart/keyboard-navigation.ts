import { useEffect } from 'react'

import { Key } from 'ts-key-enum'

import { decodePointId, getDatumValue, type SeriesDatum, type SeriesWithData } from './utils'

interface Props<Datum> {
    element: SVGSVGElement | null
    series: SeriesWithData<Datum>[]
}

export function useKeyboardNavigation<Datum>(props: Props<Datum>): void {
    const { element, series } = props

    useEffect(() => {
        if (!element) {
            return
        }

        function handleKeyPress(event: KeyboardEvent): void {
            const focusedElement = document.activeElement
            const isFocusOnTheRootElement = element === focusedElement

            if (event.key === Key.Escape) {
                element?.focus()
                return
            }

            // Focus the first element within the chart
            if (isFocusOnTheRootElement) {
                if (event.key === Key.Enter) {
                    const firstElementId = findTheFirstPointId(series)
                    const firstElement = element?.querySelector<HTMLElement>(`[data-id="${firstElementId}"]`)
                    firstElement?.focus()

                    event.preventDefault()
                    event.stopImmediatePropagation()
                }

                return
            }

            // Catch shift + tab and move focus to the root element. It prevents focusing
            // the first focusable point in case when the focus is on the second or further point
            // of the focusable point's series or any other series after it.
            if (event.shiftKey && event.key === Key.Tab) {
                event.preventDefault()
                element?.focus()
                return
            }

            if (!isArrowPressed(event)) {
                return
            }

            // Prevent native browser scrolling by arrow like key presses
            event.preventDefault()
            event.stopImmediatePropagation()

            const focusedElementId = focusedElement?.getAttribute('data-id')

            // Early exit if we can't find any focused element within the chart
            // element with special line chart id
            if (!focusedElementId) {
                return
            }

            const nextElementId = findNextElementId(event, focusedElementId, series)
            const nextElement = element?.querySelector<HTMLElement>(`[data-id="${nextElementId}"]`)

            nextElement?.focus()
        }

        element.addEventListener('keydown', handleKeyPress)

        return () => {
            element.removeEventListener('keydown', handleKeyPress)
        }
    }, [element, series])
}

function findTheFirstPointId<Datum>(series: SeriesWithData<Datum>[]): string | null {
    const sortedSeries = getSortedByFirstPointSeries(series)
    const nonEmptySeries = sortedSeries.find(series => series.data.length > 0)

    if (!nonEmptySeries) {
        return null
    }

    return nonEmptySeries.data[0].id
}

function findNextElementId<Datum>(
    event: KeyboardEvent,
    currentId: string,
    series: SeriesWithData<Datum>[]
): string | null {
    const [seriesId, index] = decodePointId(currentId)

    const sortedSeries = getSortedByFirstPointSeries(series)
    const currentSeriesIndex = sortedSeries.findIndex(series => series.id === seriesId)
    const currentSeries = sortedSeries[currentSeriesIndex]
    const currentPoint = currentSeries?.data[index]

    if (!currentSeries || !currentPoint) {
        return null
    }

    switch (event.key) {
        case Key.ArrowRight: {
            const nextPossibleIndex = index + 1

            if (nextPossibleIndex >= currentSeries.data.length) {
                const nextSeriesIndex = (currentSeriesIndex + 1) % sortedSeries.length
                const nextSeries = sortedSeries[nextSeriesIndex]

                return nextSeries.data[0].id
            }

            return currentSeries.data[nextPossibleIndex].id
        }

        case Key.ArrowLeft: {
            const nextPossibleIndex = index - 1

            if (nextPossibleIndex < 0) {
                const nextSeriesIndex = currentSeriesIndex - 1 >= 0 ? currentSeriesIndex - 1 : sortedSeries.length - 1
                const nextSeries = sortedSeries[nextSeriesIndex]

                return nextSeries.data.at(-1)!.id
            }

            return currentSeries.data[nextPossibleIndex].id
        }

        case Key.ArrowUp: {
            return getAbovePointId(currentPoint, currentSeries.id, sortedSeries)
        }

        case Key.ArrowDown: {
            return getBelowPointId(currentPoint, currentSeries.id, sortedSeries)
        }

        default: {
            return null
        }
    }
}

function getAbovePointId<Datum>(
    currentPoint: SeriesDatum<Datum>,
    currentSeriesId: string | number,
    sortedSeries: SeriesWithData<Datum>[]
): string | null {
    const currentYValue = getDatumValue(currentPoint)
    const seriesWithSameValue = getSeriesWithSameXYValue(currentPoint, sortedSeries)

    // Handle group of series with the same values case first before searching
    // for series with higher/lower value
    if (seriesWithSameValue.length > 0) {
        const currentSeriesIndex = getSeriesIndexById(currentSeriesId, seriesWithSameValue)

        // if we still within the group with same value then return next
        // series within the group
        if (currentSeriesIndex < seriesWithSameValue.length - 1) {
            const nextSeries = seriesWithSameValue[currentSeriesIndex + 1]
            return findPoint(currentPoint, nextSeries)
        }
    }

    const flatListOfAllPoints = getFlatListOfAllPoint(currentPoint, sortedSeries)

    // Try to find element above the current point
    const elementsAboveThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) > currentYValue)
        .sort(ascendingDatumOrder)

    if (elementsAboveThePoint.length > 0) {
        return elementsAboveThePoint[0].id
    }

    // Try to find element below the current point
    const elementsBelowThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) < currentYValue)
        .sort(ascendingDatumOrder)

    if (elementsBelowThePoint.length > 0) {
        return elementsBelowThePoint[0].id
    }

    // If we haven't found anything above and below the current point
    // this means there is only one case we should cover which is all series
    // are in the same point on the chart, then focus point of the first series
    // in the group
    const nextSeries = seriesWithSameValue[0]
    return findPoint(currentPoint, nextSeries)
}

function getBelowPointId<Datum>(
    currentPoint: SeriesDatum<Datum>,
    currentSeriesId: string | number,
    sortedSeries: SeriesWithData<Datum>[]
): string | null {
    const currentYValue = getDatumValue(currentPoint)
    const seriesWithSameValue = getSeriesWithSameXYValue(currentPoint, sortedSeries)

    // Handle group of series with the same values case first before searching
    // for series with higher/lower value
    if (seriesWithSameValue.length > 0) {
        const currentSeriesIndex = getSeriesIndexById(currentSeriesId, seriesWithSameValue)

        if (currentSeriesIndex > 0) {
            const nextSeries = seriesWithSameValue[currentSeriesIndex - 1]
            return findPoint(currentPoint, nextSeries)
        }
    }

    const flatListOfAllPoints = getFlatListOfAllPoint(currentPoint, sortedSeries)

    // Try to find element below the current point
    const elementsBelowThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) < currentYValue)
        .sort(descendingDatumOrder)

    if (elementsBelowThePoint.length > 0) {
        // Focus the last element within the group of series with the same values
        const lastElementFromTheBelowGroup = findLastWithSameValue(elementsBelowThePoint, item => getDatumValue(item))
        return lastElementFromTheBelowGroup?.id ?? null
    }

    // Try to find element above the current point
    const elementsAboveThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) > currentYValue)
        .sort(descendingDatumOrder)

    if (elementsAboveThePoint.length > 0) {
        // Focus the last element within the group of series with the same values
        const lastElementFromTheAboveGroup = findLastWithSameValue(elementsAboveThePoint, item => getDatumValue(item))
        return lastElementFromTheAboveGroup?.id ?? null
    }

    const nextSeries = seriesWithSameValue.at(-1)!
    return findPoint(currentPoint, nextSeries)
}

/**
 * Returns sorted series list by the first datum value in each series dataset.
 */
export function getSortedByFirstPointSeries<Datum>(series: SeriesWithData<Datum>[]): SeriesWithData<Datum>[] {
    return [...series]
        .filter(series => series.data.length > 0)
        .sort((a, b) => getDatumValue(a.data[0]) - getDatumValue(b.data[0]))
}

function findLastWithSameValue<T, D>(list: T[], mapper: (item: T) => D): T | null {
    if (list.length === 0) {
        return null
    }

    let resultElement = list[0]

    for (let index = 1; index < list.length; index++) {
        const nextValue = mapper(list[index])
        const currentValue = mapper(resultElement)

        if (currentValue !== nextValue) {
            return resultElement
        }

        resultElement = list[index]
    }

    return resultElement
}

function isArrowPressed(event: KeyboardEvent): boolean {
    return (
        event.key === Key.ArrowUp ||
        event.key === Key.ArrowRight ||
        event.key === Key.ArrowDown ||
        event.key === Key.ArrowLeft
    )
}

function ascendingDatumOrder<Datum>(a: SeriesDatum<Datum>, b: SeriesDatum<Datum>): number {
    return getDatumValue(a) - getDatumValue(b)
}

function descendingDatumOrder<Datum>(a: SeriesDatum<Datum>, b: SeriesDatum<Datum>): number {
    return getDatumValue(b) - getDatumValue(a)
}

function toInt(date: Date): number {
    return +date
}

function getSeriesWithSameXYValue<Datum>(
    point: SeriesDatum<Datum>,
    series: SeriesWithData<Datum>[]
): SeriesWithData<Datum>[] {
    return series.filter(series =>
        (series.data as SeriesDatum<Datum>[]).find(
            datum => getDatumValue(datum) === getDatumValue(point) && toInt(point.x) === toInt(datum.x)
        )
    )
}

/**
 * Returns point id within the series that has the same Y and X values
 */
function findPoint<Datum>(point: SeriesDatum<Datum>, series: SeriesWithData<Datum>): string | null {
    return (
        (series.data as SeriesDatum<Datum>[]).find(
            datum => getDatumValue(datum) === getDatumValue(point) && toInt(point.x) === toInt(datum.x)
        )?.id ?? null
    )
}

function getSeriesIndexById<Datum>(currentSeriesId: string | number, series: SeriesWithData<Datum>[]): number {
    return series.findIndex(series => series.id === currentSeriesId)
}

function getFlatListOfAllPoint<Datum>(
    point: SeriesDatum<Datum>,
    series: SeriesWithData<Datum>[]
): SeriesDatum<Datum>[] {
    return series.flatMap<SeriesDatum<Datum>>(series =>
        (series.data as SeriesDatum<Datum>[]).filter(datum => toInt(point.x) === toInt(datum.x))
    )
}

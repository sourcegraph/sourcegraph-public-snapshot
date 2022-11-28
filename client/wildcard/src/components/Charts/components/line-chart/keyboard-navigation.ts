import { useEffect } from 'react'

import { Key } from 'ts-key-enum'

import { decodePointId, getDatumValue, SeriesDatum, SeriesWithData } from './utils'

interface Props {
    element: SVGSVGElement | null
    series: SeriesWithData<any>[]
}

export function useKeyboardNavigation(props: Props): void {
    const { element, series } = props

    useEffect(() => {
        if (!element) {
            return
        }

        function handleKeyPress(event: KeyboardEvent): void {
            const focusedElement = document.activeElement
            const isFocusOnTheRootElement = element === focusedElement

            if (!isArrowPressed(event)) {
                return
            }

            // Focus the first element within the chart
            if (isFocusOnTheRootElement) {
                const firstElementId = findTheFirstPointId(series)
                const firstElement = element?.querySelector<HTMLElement>(`[data-id="${firstElementId}"]`)

                firstElement?.focus()
                return
            }

            const focusedElementId = focusedElement?.getAttribute('data-id')

            // Early exit if we can't find any focused element within the chart
            // element with special line chart id
            if (!focusedElementId) {
                return
            }

            // Prevent native browser scrolling by arrow like key presses
            event.preventDefault()
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

function findTheFirstPointId(series: SeriesWithData<unknown>[]): string | null {
    const sortedSeries = getSortedSeries(series)
    const nonEmptySeries = sortedSeries.find(series => series.data.length > 0)

    if (!nonEmptySeries) {
        return null
    }

    return nonEmptySeries.data[0].id
}

function findNextElementId(event: KeyboardEvent, currentId: string, series: SeriesWithData<any>[]): string | null {
    const [seriesId, index] = decodePointId(currentId)

    const sortedSeries = getSortedSeries(series)
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

                return nextSeries.data[nextSeries.data.length - 1].id
            }

            return currentSeries.data[nextPossibleIndex].id
        }

        case Key.ArrowUp:
            return getAbovePointId(currentPoint, currentSeries.id, sortedSeries)

        case Key.ArrowDown:
            return getBelowPointId(currentPoint, currentSeries.id, sortedSeries)

        default:
            return null
    }
}

function getAbovePointId(
    currentPoint: SeriesDatum<unknown>,
    currentSeriesId: string | number,
    sortedSeries: SeriesWithData<unknown>[]
): string | null {
    const currentYValue = getDatumValue(currentPoint)

    const seriesWithSameValue = sortedSeries.filter(series =>
        (series.data as SeriesDatum<any>[]).find(
            datum => getDatumValue(datum) === currentYValue && +currentPoint.x === +datum.x
        )
    )

    if (seriesWithSameValue.length > 0) {
        const currentSeriesIndex = seriesWithSameValue.findIndex(series => series.id === currentSeriesId)

        if (currentSeriesIndex < seriesWithSameValue.length - 1) {
            const nextSeries = seriesWithSameValue[currentSeriesIndex + 1]
            return (
                (nextSeries.data as SeriesDatum<unknown>[]).find(
                    datum => getDatumValue(datum) === currentYValue && +currentPoint.x === +datum.x
                )?.id ?? null
            )
        }
    }

    const flatListOfAllPoints = sortedSeries.flatMap<SeriesDatum<unknown>>(series =>
        (series.data as SeriesDatum<unknown>[]).filter(datum => +currentPoint.x === +datum.x)
    )

    const elementsAboveThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) > currentYValue)
        .sort((a, b) => getDatumValue(a) - getDatumValue(b))

    if (elementsAboveThePoint.length > 0) {
        return elementsAboveThePoint[0].id
    }

    const elementsBelowThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) < currentYValue)
        .sort((a, b) => getDatumValue(a) - getDatumValue(b))

    if (elementsBelowThePoint.length > 0) {
        return elementsBelowThePoint[0].id
    }

    const nextSeries = seriesWithSameValue[0]

    return (
        (nextSeries.data as SeriesDatum<unknown>[]).find(
            datum => getDatumValue(datum) === currentYValue && +currentPoint.x === +datum.x
        )?.id ?? null
    )
}

function getBelowPointId(
    currentPoint: SeriesDatum<unknown>,
    currentSeriesId: string | number,
    sortedSeries: SeriesWithData<unknown>[]
): string | null {
    const currentYValue = getDatumValue(currentPoint)

    const seriesWithSameValue = sortedSeries.filter(series =>
        (series.data as SeriesDatum<any>[]).find(
            datum => getDatumValue(datum) === currentYValue && +currentPoint.x === +datum.x
        )
    )

    if (seriesWithSameValue.length > 0) {
        const currentSeriesIndex = seriesWithSameValue.findIndex(series => series.id === currentSeriesId)

        if (currentSeriesIndex > 0) {
            const nextSeries = seriesWithSameValue[currentSeriesIndex - 1]

            return (
                (nextSeries.data as SeriesDatum<unknown>[]).find(
                    datum => getDatumValue(datum) === currentYValue && +currentPoint.x === +datum.x
                )?.id ?? null
            )
        }
    }

    const flatListOfAllPoints = sortedSeries.flatMap<SeriesDatum<any>>(series =>
        (series.data as SeriesDatum<unknown>[]).filter(datum => +currentPoint.x === +datum.x)
    )

    const elementsBelowThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) < currentYValue)
        .sort((a, b) => getDatumValue(b) - getDatumValue(a))

    if (elementsBelowThePoint.length > 0) {
        const lastElementFromTheBelowGroup = findLastWithSameValue(elementsBelowThePoint, item => getDatumValue(item))
        return lastElementFromTheBelowGroup?.id ?? null
    }

    const elementsAboveThePoint = flatListOfAllPoints
        .filter(datum => getDatumValue(datum) > currentYValue)
        .sort((a, b) => getDatumValue(b) - getDatumValue(a))

    if (elementsAboveThePoint.length > 0) {
        const lastElementFromTheAboveGroup = findLastWithSameValue(elementsAboveThePoint, item => getDatumValue(item))
        return lastElementFromTheAboveGroup?.id ?? null
    }

    const nextSeries = seriesWithSameValue[seriesWithSameValue.length - 1]

    return (
        (nextSeries.data as SeriesDatum<unknown>[]).find(
            datum => getDatumValue(datum) === currentYValue && +currentPoint.x === +datum.x
        )?.id ?? null
    )
}

/**
 * Returns sorted series list by the first datum value in each series dataset.
 */
function getSortedSeries(series: SeriesWithData<any>[]): SeriesWithData<any>[] {
    return [...series].sort((a, b) => getDatumValue(a.data[0]) - getDatumValue(b.data[0]))
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

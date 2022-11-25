import { useLayoutEffect, useMemo, useState } from 'react'

export type Measurements = Pick<
    DOMRectReadOnly,
    'x' | 'y' | 'top' | 'left' | 'right' | 'bottom' | 'height' | 'width'
> & {
    parentWidth: number
    parentHeight: number
}

const DEFAULTS: Measurements = {
    x: 0,
    y: 0,
    width: 0,
    height: 0,
    top: 0,
    left: 0,
    bottom: 0,
    right: 0,
    parentWidth: 0,
    parentHeight: 0,
}

type MeasurementType = 'contentRect' | 'boundingRect'

/**
 * Custom hook for observing the size and position of an element using the Resize Observer
 * API. Based on https://github.com/streamich/react-use/blob/master/src/useMeasure.ts
 */
export const useMeasure = <E extends Element = Element>(
    element?: Element | null,
    type: MeasurementType = 'contentRect',
    includeParentDimension: boolean = false
): [ref: (element: E | null) => void, measurements: Measurements] => {
    const [internalElement, setInternalElement] = useState<E | null>(null)
    const [measurements, setMeasurements] = useState<Measurements>(DEFAULTS)
    const elementToMeasure = element ?? internalElement

    const observer = useMemo(
        () =>
            new window.ResizeObserver(entries => {
                if (entries[0]) {
                    const rect =
                        type === 'contentRect' ? entries[0].contentRect : entries[0].target.getBoundingClientRect()

                    const { x, y, width, height, top, left, bottom, right } = rect

                    let parentWidth = 0
                    let parentHeight = 0
                    if (includeParentDimension) {
                        const parentRect = entries[0].target.parentElement?.getBoundingClientRect()
                        parentWidth = parentRect?.width ?? 0
                        parentHeight = parentRect?.height ?? 0
                    }

                    setMeasurements({ x, y, width, height, top, left, bottom, right, parentWidth, parentHeight })
                }
            }),
        [type, includeParentDimension]
    )

    useLayoutEffect(() => {
        if (!elementToMeasure) {
            return
        }
        observer.observe(elementToMeasure)

        return () => observer.disconnect()
    }, [elementToMeasure, observer])

    return [setInternalElement, measurements]
}

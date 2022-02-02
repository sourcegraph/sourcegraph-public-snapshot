import { useLayoutEffect, useMemo, useState } from 'react'

export type Measurements = Pick<DOMRectReadOnly, 'x' | 'y' | 'top' | 'left' | 'right' | 'bottom' | 'height' | 'width'>

const DEFAULTS: Measurements = {
    x: 0,
    y: 0,
    width: 0,
    height: 0,
    top: 0,
    left: 0,
    bottom: 0,
    right: 0,
}

/**
 * Custom hook for observing the size and position of an element using the Resize Observer
 * API. Based on https://github.com/streamich/react-use/blob/master/src/useMeasure.ts
 */
export const useMeasure = <E extends Element = Element>(): [
    ref: (element: E | null) => void,
    measurements: Measurements
] => {
    const [element, setElement] = useState<E | null>(null)
    const [measurements, setMeasurements] = useState<Measurements>(DEFAULTS)

    const observer = useMemo(
        () =>
            new window.ResizeObserver(entries => {
                if (entries[0]) {
                    // eslint-disable-next-line id-length
                    const { x, y, width, height, top, left, bottom, right } = entries[0].contentRect
                    setMeasurements({ x, y, width, height, top, left, bottom, right })
                }
            }),
        []
    )

    useLayoutEffect(() => {
        if (!element) {
            return
        }
        observer.observe(element)

        return () => observer.disconnect()
    }, [element, observer])

    return [setElement, measurements]
}

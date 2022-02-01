import { Accessors } from '../types'

/** Returns minimal and maximal value from data series */
export function getMinAndMax<Datum, Key extends keyof Datum>(
    data: Datum[],
    accessors: Accessors<Datum, Key>
): [number, number] {
    const keys = Object.keys(accessors.y) as Key[]

    const resultArray = data.reduce<number[]>((memo, item) => {
        for (const key of keys) {
            const accessor = accessors.y[key]

            memo.push(+accessor(item))
        }

        return memo
    }, [])

    return [Math.min(...resultArray), Math.max(...resultArray)]
}

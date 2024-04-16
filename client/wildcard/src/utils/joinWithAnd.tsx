import React from 'react'

/**
 * Returns a ReactNode where all the items in the provided array have been visually joined
 * with commas, and an "and" is inserted before the last item.
 *
 * @param array The array to format as a list with commas and an "and" before the last
 * item
 * @param formatItem Function to call to format each item in the array before joining
 * @param getKey Function to call to get a unique key for each item in the array
 * @param maxItems Optionally, the maximum number of items to show. If the array has more
 * items than this, the last item will be followed with "and N more".
 */
export function joinWithAnd<T>(
    array: T[],
    formatItem: (original: T) => React.ReactNode,
    getKey: (original: T) => string,
    maxItems?: number
): React.ReactNode {
    if (array.length === 0) {
        return null
    }
    if (array.length <= 1) {
        return formatItem(array[0])
    }
    if (array.length === 2) {
        return (
            <>
                {formatItem(array[0])} and {formatItem(array[1])}
            </>
        )
    }
    if (maxItems && array.length > maxItems) {
        return (
            <>
                {array.slice(0, maxItems).map(item => (
                    <React.Fragment key={getKey(item)}>{formatItem(item)}, </React.Fragment>
                ))}
                and {array.length - maxItems} more
            </>
        )
    }
    return (
        <>
            {array.slice(0, -1).map(item => (
                <React.Fragment key={getKey(item)}>{formatItem(item)}, </React.Fragment>
            ))}
            and {formatItem(array.at(-1)!)}
        </>
    )
}

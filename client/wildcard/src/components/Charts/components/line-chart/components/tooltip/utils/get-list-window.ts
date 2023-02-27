export interface ListWindow<T> {
    window: T[]
    leftRemaining: number
    rightRemaining: number
}

/**
 * Return window (sub-list) around point with {@link index} and with size
 * equals to {@link size}.
 *
 * Example:
 * list - 1,2,3,4,5,6,7 index - 3, size - 3
 * result 3,4,5 because 1,2, * 3, 4, 5 * 6, 7
 * remaining left 2 (1,2) remaining right 2 (6, 7)
 */
export function getListWindow<T>(list: T[], index: number, size: number): ListWindow<T> {
    if (list.length < size) {
        return { window: list, leftRemaining: 0, rightRemaining: 0 }
    }

    let left = index
    let right = index
    const window = [list[index]]

    while (window.length < size) {
        const nextLeft = left - 1
        const leftElement = list[nextLeft]

        if (leftElement) {
            left--
            window.unshift(leftElement)
        }

        const nextRight = right + 1
        const rightElement = list[nextRight]

        if (rightElement) {
            right++
            window.push(rightElement)
        }

        if (!leftElement && !rightElement) {
            break
        }
    }

    return {
        window,
        leftRemaining: Math.max(left, 0),
        rightRemaining: Math.max(list.length - 1 - right, 0),
    }
}

/**
 * Helper replace element in array by index and return new array.
 * */
export function replace<Element>(list: Element[], index: number, newElement: Element): Element[] {
    return [...list.slice(0, index), newElement, ...list.slice(index + 1)]
}

/**
 * Helper remove element from array by index
 * */
export function remove<Element>(list: Element[], index: number): Element[] {
    return [...list.slice(0, index), ...list.slice(index + 1)]
}

function bubbleSort<T>(items: T[]): T[] {
    var length = items.length
    for (var i = 0; i < length; i++) {
        // Last i elements are already in place
        // for next iteration we need to shift

        // traverse the array from 0 to length - i - 1
        // Swap if the element found is greater than the next element
        // Move to next element for next iteration

        for (var j = 0; j < length - i - 1; j++) {
            if (items[j] > items[j + 1]) {
                var tmp = items[j]
                items[j] = items[j + 1]
                items[j + 1] = tmp
            }
        }
    }
    return items
}

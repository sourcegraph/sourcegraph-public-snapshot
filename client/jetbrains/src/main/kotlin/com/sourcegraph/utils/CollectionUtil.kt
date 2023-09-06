package com.sourcegraph.utils

class CollectionUtil {
    companion object {
        infix fun <T> List<T>.sameElements(anotherList: List<T>) =
            this.size == anotherList.size && this.toSet() == anotherList.toSet()
        infix fun <T> List<T>.diff(anotherList: List<T>) =
            this.minus(anotherList.toSet())
    }
}

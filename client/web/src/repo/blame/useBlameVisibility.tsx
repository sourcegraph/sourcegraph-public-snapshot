import { useCallback } from 'react'

import { BehaviorSubject } from 'rxjs'

import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

const IS_BLAME_VISIBLE_STORAGE_KEY = 'GitBlame.isVisible'
const isBlameVisibleObservable = new BehaviorSubject<boolean | undefined>(undefined)

export const useBlameVisibility = (): [boolean, (isVisible: boolean) => void] => {
    const [isVisibleFromLocalStorage, updateLocalStorageValue] = useLocalStorage(IS_BLAME_VISIBLE_STORAGE_KEY, false)
    const isVisible = useObservable(isBlameVisibleObservable)
    const setIsBlameVisible = useCallback(
        (isVisible: boolean): void => {
            isBlameVisibleObservable.next(isVisible)
            updateLocalStorageValue(isVisible)
        },
        [updateLocalStorageValue]
    )

    return [isVisible ?? isVisibleFromLocalStorage, setIsBlameVisible]
}

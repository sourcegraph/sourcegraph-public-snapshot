import { useCallback } from 'react'

import { BehaviorSubject } from 'rxjs'

import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

const IS_BLAME_VISIBLE_STORAGE_KEY = 'GitBlame.isVisible'
const isBlameVisible = new BehaviorSubject<boolean | undefined>(undefined)

export const useBlameVisibility = (isPackage: boolean): [boolean, (isVisible: boolean) => void] => {
    const [isVisibleFromLocalStorage, updateLocalStorageValue] = useLocalStorage(IS_BLAME_VISIBLE_STORAGE_KEY, false)
    const isVisibleFromObservable = useObservable(isBlameVisible)
    const setIsBlameVisible = useCallback(
        (isVisible: boolean): void => {
            isBlameVisible.next(isVisible)
            updateLocalStorageValue(isVisible)
        },
        [updateLocalStorageValue]
    )

    return [!isPackage && (isVisibleFromObservable ?? isVisibleFromLocalStorage), setIsBlameVisible]
}

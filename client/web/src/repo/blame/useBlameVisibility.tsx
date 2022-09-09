import { BehaviorSubject } from 'rxjs'

import { useObservable } from '@sourcegraph/wildcard'

const IS_BLAME_VISIBLE_STORAGE_KEY = 'GitBlame.isVisible'
const isBlameVisible = new BehaviorSubject<boolean>(!!window.localStorage.getItem(IS_BLAME_VISIBLE_STORAGE_KEY))
const setIsBlameVisible = (isVisible: boolean): void => {
    isBlameVisible.next(isVisible)

    if (isVisible) {
        window.localStorage.setItem(IS_BLAME_VISIBLE_STORAGE_KEY, 'true')
    } else {
        window.localStorage.removeItem(IS_BLAME_VISIBLE_STORAGE_KEY)
    }
}

export const useBlameVisibility = (): [boolean, (isVisible: boolean) => void] => {
    const isVisible = useObservable(isBlameVisible)

    return [Boolean(isVisible), setIsBlameVisible]
}

import { BehaviorSubject } from 'rxjs'

import { useObservable } from '@sourcegraph/wildcard'

const isBlameVisible = new BehaviorSubject<boolean>(false)
const setIsBlameVisible = (isVisible: boolean): void => isBlameVisible.next(isVisible)

export const useBlameVisibility = (): [boolean, (isVisible: boolean) => void] => {
    const isVisible = useObservable(isBlameVisible)

    return [Boolean(isVisible), setIsBlameVisible]
}

import { isEqual } from 'lodash'
import { fromEvent, type Observable } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'

import { lprToSelectionsZeroIndexed } from '@sourcegraph/common'
import type { Selection } from '@sourcegraph/extension-api-types'
import { parseHash } from '@sourcegraph/shared/src/util/url'

export function getSelectionsFromHash(): Selection[] {
    return lprToSelectionsZeroIndexed(parseHash(window.location.hash))
}

export function observeSelectionsFromHash(): Observable<Selection[]> {
    return fromEvent(window, 'hashchange').pipe(
        map(getSelectionsFromHash),
        distinctUntilChanged((a, b) => isEqual(a, b))
    )
}

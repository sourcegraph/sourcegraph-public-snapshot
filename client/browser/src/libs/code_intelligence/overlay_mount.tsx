import { Observable } from 'rxjs'
import { mapTo, tap } from 'rxjs/operators'

/** Emits when the overlay mount is removed from the DOM. */
export const onMountRemovedFromDOM: () => (mounts: Observable<HTMLElement>) => Observable<void> = () => {
    const domObserver = new MutationObserver(mutations => {
        for (const mutation of mutations) {
            console.log(mutation)
        }
    })

    return mounts =>
        mounts.pipe(
            tap(mount => {
                domObserver.observe(mount)
            }),
            mapTo(undefined)
        )
}

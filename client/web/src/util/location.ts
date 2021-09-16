import * as H from 'history'
import { Observable } from 'rxjs'

export function observeLocation(history: H.History): Observable<H.Location> {
    return new Observable(function subscribe(subscriber) {
        const unlisten = history.listen(location => {
            subscriber.next(location)
        })

        return unlisten
    })
}

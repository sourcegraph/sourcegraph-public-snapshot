import { concat, fromEvent, Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

export const observeMediaQuery = (query: string): Observable<boolean> => {
    const mediaList = window.matchMedia(query)
    return concat(
        of(mediaList.matches),
        fromEvent<MediaQueryListEvent>(mediaList, 'change').pipe(map(event => event.matches))
    )
}

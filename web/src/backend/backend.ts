import { Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { fromFetch } from 'rxjs/fetch'
import { checkOk } from '../../../shared/src/backend/fetch'

export function backendRequest<T>(url: string): Observable<T> {
    return fromFetch(url, { headers: window.context.xhrHeaders }).pipe(
        map(checkOk),
        switchMap(response => response.json())
    )
}

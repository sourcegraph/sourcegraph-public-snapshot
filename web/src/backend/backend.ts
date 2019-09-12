import { ajax, AjaxResponse } from 'rxjs/ajax'
import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { normalizeAjaxError } from '../../../shared/src/util/errors'

export function backendRequest<T>(url: string): Observable<T> {
    return ajax({
        url,
        headers: window.context.xhrHeaders,
    }).pipe(
        catchError<AjaxResponse, never>(err => {
            normalizeAjaxError(err)
            throw err
        }),
        map<AjaxResponse, T>(({ response }) => response)
    )
}

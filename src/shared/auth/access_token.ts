import { omit } from 'lodash'
import { Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import storage from '../../browser/storage'

export const getAccessToken = (url: string): Observable<string | undefined> =>
    new Observable(observer => {
        storage.getSync(items => {
            observer.next(items.accessTokens[url])
            observer.complete()
        })
    })

export const setAccessToken = (url: string) => (tokens: Observable<string>): Observable<string> =>
    tokens.pipe(
        switchMap(
            token =>
                new Observable<string>(observer => {
                    storage.getSync(({ accessTokens }) =>
                        storage.setSync({ accessTokens: { ...accessTokens, [url]: token } }, () => {
                            observer.next(token)
                            observer.complete()
                        })
                    )
                })
        )
    )

export const removeAccessToken = (url: string): Observable<void> =>
    new Observable(observer => {
        storage.getSync(({ accessTokens }) =>
            storage.setSync({ accessTokens: omit(accessTokens, url) }, () => {
                observer.next()
                observer.complete()
            })
        )
    })

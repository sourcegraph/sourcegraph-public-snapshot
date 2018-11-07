import { omit } from 'lodash'
import { Observable } from 'rxjs'
import storage from '../../browser/storage'
import { AccessToken } from '../../browser/types'

export const getAccessToken = (url: string): Observable<AccessToken | undefined> =>
    new Observable(observer => {
        storage.getSync(items => {
            observer.next(items.accessTokens[url])
            observer.complete()
        })
    })

export const setAccessToken = (url: string, token: AccessToken) =>
    storage.getSync(({ accessTokens }) => storage.setSync({ accessTokens: { ...accessTokens, [url]: token } }))

export const removeAccessToken = (url: string): Observable<void> =>
    new Observable(observer => {
        storage.getSync(({ accessTokens }) =>
            storage.setSync({ accessTokens: omit(accessTokens, url) }, () => {
                observer.next()
                observer.complete()
            })
        )
    })

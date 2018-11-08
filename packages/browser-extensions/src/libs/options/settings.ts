import { pick } from 'lodash'
import { Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import storage from '../../browser/storage'
import { configurableFeatureFlags, FeatureFlags } from '../../browser/types'

const configurableKeys = Object.keys(configurableFeatureFlags).filter(key => configurableFeatureFlags[key])

const getConfigurableSettingsFromObject = (obj: any): Partial<FeatureFlags> =>
    pick(obj, configurableKeys) as Partial<FeatureFlags>

export const getConfigurableSettings = (): Observable<Partial<FeatureFlags>> =>
    storage.observeSync('featureFlags').pipe(map(featureFlags => getConfigurableSettingsFromObject(featureFlags)))

export const setConfigurabelSettings = (settings: Partial<FeatureFlags>): Observable<Partial<FeatureFlags>> =>
    new Observable<void>(observer =>
        storage.getSync(items => {
            storage.setSync(
                {
                    featureFlags: {
                        ...items.featureFlags,
                        ...settings,
                    },
                },
                () => {
                    observer.next()
                    observer.complete()
                }
            )
        })
    ).pipe(switchMap(() => getConfigurableSettings()))

export const setConfigurabelSettingsPromise = (settings: Partial<FeatureFlags>): Promise<Partial<FeatureFlags>> =>
    new Promise<Partial<FeatureFlags>>(resolve => {
        storage.getSync(items => {
            resolve(items.featureFlags)
        })
    }).then(
        featureFlags =>
            new Promise<Partial<FeatureFlags>>(resolve => {
                const value = { ...featureFlags, ...getConfigurableSettingsFromObject(settings) } as FeatureFlags

                storage.setSync({ featureFlags: value }, () => resolve(value))
            })
    )

export const setSourcegraphURL = (sourcegraphURL: string) => {
    storage.setSync({ sourcegraphURL })
}

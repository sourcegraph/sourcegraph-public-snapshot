import { Observable, ReplaySubject, Subject } from 'rxjs'
import { share, switchMap, take } from 'rxjs/operators'
import { StorageItems } from './types'

type MigrateFunc = (items: StorageItems) => { newItems?: StorageItems; keysToRemove?: string[] }

export interface MigratableStorageArea extends chrome.storage.StorageArea {
    setMigration: (migrate: MigrateFunc) => void
}

export const noopMigration: MigrateFunc = () => ({})

export function provideMigrations(area: chrome.storage.StorageArea): MigratableStorageArea {
    const migrations = new Subject<MigrateFunc>()
    const getCalls = new ReplaySubject<any[]>()
    const setCalls = new ReplaySubject<any[]>()

    const migrated = migrations.pipe(
        switchMap(
            migrate =>
                new Observable<void>(observer => {
                    area.get(items => {
                        const { newItems, keysToRemove } = migrate(items as StorageItems)
                        area.remove(keysToRemove || [], () => {
                            area.set(newItems || {}, () => {
                                observer.next()
                                observer.complete()
                            })
                        })
                    })
                })
        ),
        take(1),
        share()
    )

    const initializedGets = migrated.pipe(switchMap(() => getCalls))
    const initializedSets = migrated.pipe(switchMap(() => setCalls))

    initializedSets.subscribe(args => {
        area.set.apply(area, args)
    })

    initializedGets.subscribe(args => {
        area.get.apply(area, args)
    })

    const get: chrome.storage.StorageArea['get'] = (...args) => {
        getCalls.next(args)
    }

    const set: chrome.storage.StorageArea['set'] = (...args) => {
        setCalls.next(args)
    }

    return {
        ...area,
        get,
        set,

        setMigration: migrate => {
            migrations.next(migrate)
        },
    }
}

import { applyEdits, modify, parse as parseJSONC } from 'jsonc-parser'
import { from, fromEvent, Observable } from 'rxjs'
import { distinctUntilChanged, filter, map, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { SettingsEdit } from '@sourcegraph/shared/src/api/client/services/settings'
import { viewerSettingsQuery } from '@sourcegraph/shared/src/backend/settings'
import { ViewerSettingsResult } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { ISettingsCascade } from '@sourcegraph/shared/src/schema'
import {
    mergeSettings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { observeStorageKey, storage } from '../../browser-extension/web-extension-api/storage'
import { isInPage } from '../context'

const inPageClientSettingsKey = 'sourcegraphClientSettings'

/**
 * Returns an observable that emits the localStorage value (as a raw string) for the given key on
 * every storage update event, starting with the current value.
 */
function observeLocalStorageKey(key: string, defaultValue: string): Observable<string> {
    const initialValue = localStorage.getItem(key)
    return fromEvent<StorageEvent>(window, 'storage').pipe(
        filter(event => event.key === key),
        map(event => event.newValue),
        startWith(initialValue),
        map(value => value ?? defaultValue),
        distinctUntilChanged()
    )
}

const createStorageSettingsCascade: () => Observable<SettingsCascade> = () => {
    /** Observable of the JSONC string of the settings.
     *
     * NOTE: We can't use LocalStorageSubject here because the JSONC string is stored raw in localStorage and LocalStorageSubject also does parsing.
     * This could be changed, but users already have settings stored, so it would need a migration for little benefit.
     */
    const storageObservable = isInPage
        ? observeLocalStorageKey(inPageClientSettingsKey, '{}')
        : observeStorageKey('sync', 'clientSettings')

    const subject: SettingsSubject = {
        __typename: 'Client',
        id: 'Client',
        displayName: 'Client',
        viewerCanAdminister: true,
    }

    return storageObservable.pipe(
        map(clientSettingsString => parseJSONC(clientSettingsString || '')),
        map(clientSettings => ({
            subjects: [
                {
                    subject,
                    settings: clientSettings,
                    lastID: null,
                },
            ],
            final: clientSettings || {},
        }))
    )
}

/**
 * The settings cascade consisting solely of client settings.
 */
export const storageSettingsCascade = createStorageSettingsCascade()

/**
 * Merge two settings cascades (used to merge viewer settings and client settings).
 */
export function mergeCascades(
    cascadeOrError: SettingsCascadeOrError,
    cascade: SettingsCascade
): SettingsCascadeOrError {
    return {
        subjects:
            cascadeOrError.subjects === null
                ? cascade.subjects
                : isErrorLike(cascadeOrError.subjects)
                ? cascadeOrError.subjects
                : [...cascadeOrError.subjects, ...cascade.subjects],
        final:
            cascadeOrError.final === null
                ? cascade.final
                : isErrorLike(cascadeOrError.final)
                ? cascadeOrError.final
                : mergeSettings([cascadeOrError.final, cascade.final]),
    }
}

/**
 * Fetches the settings cascade for the viewer.
 */
export function fetchViewerSettings(requestGraphQL: PlatformContext['requestGraphQL']): Observable<ISettingsCascade> {
    return from(
        requestGraphQL<ViewerSettingsResult>({
            request: viewerSettingsQuery,
            variables: {},
            mightContainPrivateInfo: false,
        })
    ).pipe(
        map(dataOrThrowErrors),
        map(({ viewerSettings }) => {
            if (!viewerSettings) {
                throw new Error('fetchViewerSettings: empty viewerSettings')
            }

            for (const subject of viewerSettings.subjects) {
                // User/org/global settings cannot be edited from the
                // browser extension (only client settings can).
                subject.viewerCanAdminister = false
            }

            return viewerSettings as ISettingsCascade
        })
    )
}

/**
 * Applies an edit and persists the result to client settings.
 */
export async function editClientSettings(edit: SettingsEdit | string): Promise<void> {
    const getNext = (previous: string): string =>
        typeof edit === 'string'
            ? edit
            : applyEdits(
                  previous,
                  modify(previous, edit.path, edit.value, {
                      formattingOptions: {
                          tabSize: 2,
                          insertSpaces: true,
                          eol: '\n',
                      },
                  })
              )
    if (isInPage) {
        const previous = localStorage.getItem(inPageClientSettingsKey) || ''
        const next = getNext(previous)

        localStorage.setItem(inPageClientSettingsKey, next)

        return Promise.resolve()
    }

    const { clientSettings: previous = '{}' } = await storage.sync.get()
    const next = getNext(previous)

    await storage.sync.set({ clientSettings: next })
}

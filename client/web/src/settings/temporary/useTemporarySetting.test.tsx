import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { renderHook, act } from '@testing-library/react-hooks'
import React from 'react'
import { Observable, of } from 'rxjs'
import { delay } from 'rxjs/operators'

import { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsContext } from './TemporarySettingsProvider'
import { SettingsBackend, TemporarySettingsStorage } from './TemporarySettingsStorage'
import { useTemporarySetting } from './useTemporarySetting'

describe('useTemporarySetting', () => {
    class InMemoryMockSettingsBackend implements SettingsBackend {
        constructor(private settings: TemporarySettings) {}
        public load(): Observable<TemporarySettings> {
            return new Observable(subscribe => {
                setTimeout(() => {
                    subscribe.next(this.settings)
                    subscribe.complete()
                }, 100)
            })
        }
        public save(settings: TemporarySettings): Observable<void> {
            this.settings = settings
            return of()
        }
    }

    const mockClient = createMockClient(
        null,
        gql`
            query {
                temporarySettings {
                    contents
                }
            }
        `
    )

    it('should get correct data from storage', async () => {
        const settingsBackend = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { filters: true, reference: false },
        })
        const settingsStorage = new TemporarySettingsStorage(mockClient, null)
        settingsStorage.setSettingsBackend(settingsBackend)

        const { result, waitFor } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })
        await waitFor(() => !!result.current[0])
        const [value] = result.current
        expect(value).toEqual({ filters: true, reference: false })
    })

    it('should get undefined if data does not exist in storage', async () => {
        const settingsBackend = new InMemoryMockSettingsBackend({})
        const settingsStorage = new TemporarySettingsStorage(mockClient, null)
        settingsStorage.setSettingsBackend(settingsBackend)

        const { result, waitFor } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })

        await waitFor(() => !!result.current[0])
        const [value] = result.current
        expect(value).toBe(undefined)
    })

    it('should save data and update value', async () => {
        const settingsBackend = new InMemoryMockSettingsBackend({})
        const settingsStorage = new TemporarySettingsStorage(mockClient, null)
        settingsStorage.setSettingsBackend(settingsBackend)

        const { result } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })

        const [, setValue] = result.current
        act(() => setValue({ filters: true, reference: false }))

        const [value] = result.current
        expect(value).toEqual({ filters: true, reference: false })
    })

    it('should update other hook values if changed in another hook', async () => {
        const settingsBackend = new InMemoryMockSettingsBackend({})
        const settingsStorage = new TemporarySettingsStorage(mockClient, null)
        settingsStorage.setSettingsBackend(settingsBackend)

        const { result: result1 } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })

        const { result: result2 } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })

        const [, setValue] = result1.current
        act(() => setValue({ filters: true, reference: false }))

        const [value] = result2.current
        expect(value).toEqual({ filters: true, reference: false })
    })

    it('should update data if backend changed', async () => {
        const settingsBackend1 = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { filters: true, reference: false },
        })
        const settingsStorage = new TemporarySettingsStorage(mockClient, null)
        settingsStorage.setSettingsBackend(settingsBackend1)

        const { result } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })

        const settingsBackend2 = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { repositories: true },
        })

        act(() => settingsStorage.setSettingsBackend(settingsBackend2))

        const [value] = result.current
        expect(value).toEqual({ repositories: true })
    })
})

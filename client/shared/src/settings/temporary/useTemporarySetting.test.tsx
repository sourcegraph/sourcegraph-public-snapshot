import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { renderHook, act } from '@testing-library/react-hooks'
import React from 'react'

import { TemporarySettingsContext } from './TemporarySettingsProvider'
import { InMemoryMockSettingsBackend, TemporarySettingsStorage } from './TemporarySettingsStorage'
import { useTemporarySetting } from './useTemporarySetting'

describe('useTemporarySetting', () => {
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

    it('should get correct data from storage', () => {
        const settingsBackend = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { filters: true, reference: false },
        })
        const settingsStorage = new TemporarySettingsStorage(mockClient, false)
        settingsStorage.setSettingsBackend(settingsBackend)

        const { result } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })
        const [value] = result.current
        expect(value).toEqual({ filters: true, reference: false })
    })

    it('should get undefined if data does not exist in storage', () => {
        const settingsBackend = new InMemoryMockSettingsBackend({})
        const settingsStorage = new TemporarySettingsStorage(mockClient, false)
        settingsStorage.setSettingsBackend(settingsBackend)

        const { result } = renderHook(() => useTemporarySetting('search.collapsedSidebarSections'), {
            wrapper: ({ children }) => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    {children}
                </TemporarySettingsContext.Provider>
            ),
        })

        const [value] = result.current
        expect(value).toBe(undefined)
    })

    it('should save data and update value', () => {
        const settingsBackend = new InMemoryMockSettingsBackend({})
        const settingsStorage = new TemporarySettingsStorage(mockClient, false)
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
        act(() => setValue({ filters: true, reference: false }))

        const [value] = result.current
        expect(value).toEqual({ filters: true, reference: false })
    })

    it('should update other hook values if changed in another hook', () => {
        const settingsBackend = new InMemoryMockSettingsBackend({})
        const settingsStorage = new TemporarySettingsStorage(mockClient, false)
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

    it('should update data if backend changed', () => {
        const settingsBackend1 = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { filters: true, reference: false },
        })
        const settingsStorage = new TemporarySettingsStorage(mockClient, false)
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

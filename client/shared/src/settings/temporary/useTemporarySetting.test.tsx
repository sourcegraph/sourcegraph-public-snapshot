import { useEffect } from 'react'

import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { render, renderHook, act as actHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { TemporarySettingsContext } from './TemporarySettingsProvider'
import { InMemoryMockSettingsBackend, TemporarySettingsStorage } from './TemporarySettingsStorage'
import { useTemporarySetting } from './useTemporarySetting'

describe('useTemporarySetting', () => {
    const mockClient = createMockClient(
        null,
        gql`
            query TemporarySettings {
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
        actHook(() => setValue({ filters: true, reference: false }))
        actHook(() => setValue({ filters: true, reference: false }))

        const [value] = result.current
        expect(value).toEqual({ filters: true, reference: false })
    })

    it('should support the updater callback pattern', () => {
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
        actHook(() =>
            setValue(currentValue => {
                expect(currentValue).toEqual(undefined)
                return { filters: true, reference: false }
            })
        )
        actHook(() =>
            setValue(currentValue => {
                expect(currentValue).toEqual({ filters: true, reference: false })
                return { filters: true, reference: true }
            })
        )

        const [value] = result.current
        expect(value).toEqual({ filters: true, reference: true })
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
        actHook(() => setValue({ filters: true, reference: false }))

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

        actHook(() => settingsStorage.setSettingsBackend(settingsBackend2))

        const [value] = result.current
        expect(value).toEqual({ repositories: true })
    })

    it('should not recreate the updater function', () => {
        const settingsBackend1 = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { filters: true, reference: false },
        })
        const settingsStorage = new TemporarySettingsStorage(mockClient, false)
        settingsStorage.setSettingsBackend(settingsBackend1)

        let updateCount = 0
        const Component = () => {
            const [_setting, setSetting] = useTemporarySetting('search.collapsedSidebarSections')
            useEffect(() => {
                updateCount++
            }, [setSetting])
            return null
        }

        render(
            <TemporarySettingsContext.Provider value={settingsStorage}>
                <Component />
            </TemporarySettingsContext.Provider>
        )

        expect(updateCount).toBe(1)
    })

    it('should always provides latest previousValue', () => {
        const settingsBackend = new InMemoryMockSettingsBackend({
            'search.collapsedSidebarSections': { filters: true, reference: true },
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

        expect(result.current[0]).toEqual({ filters: true, reference: true })
        actHook(() => {
            result.current[1](previousValue => ({
                ...previousValue,
                filters: false,
            }))
            result.current[1](previousValue => ({
                ...previousValue,
                reference: false,
            }))
        })
        expect(result.current[0]).toEqual({ filters: false, reference: false })
    })
})

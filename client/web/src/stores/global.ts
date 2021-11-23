import create, { SetState, StoreApi } from 'zustand'
import { persist } from 'zustand/middleware'

import { NavbarQueryState, createNavbarQueryStateStore } from './navbarSearchQueryState'
import {
    createThemeStateStore,
    LIGHT_THEME_LOCAL_STORAGE_KEY,
    readStoredThemePreference,
    ThemeState,
} from './themeState'

interface GlobalStore extends NavbarQueryState, ThemeState {}

/**
 * This store is used to separate global/shared state from our component hierarchy.
 *
 * We want to keep shared state in a single place. If you need to share data
 * across various components in the application, update this store instead of
 * creating a separate one.
 *
 * Note: We are in the process of migrating shared data to this store, so not
 * everything that should be in here is already in here.
 */
export const useGlobalStore = create(
    persist<GlobalStore>(
        (set, get, api) => ({
            ...createNavbarQueryStateStore(set as SetState<any>, get, api as StoreApi<any>),
            ...createThemeStateStore(set as SetState<any>),
        }),
        {
            name: LIGHT_THEME_LOCAL_STORAGE_KEY,
            whitelist: ['theme'],
            serialize: state => state.state.theme,
            deserialize: string => ({ state: { theme: readStoredThemePreference(string) } }),
        }
    )
)

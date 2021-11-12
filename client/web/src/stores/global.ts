import create from 'zustand'

import { NavbarQueryState, createNavbarQueryStateStore } from './navbarSearchQueryState'

interface GlobalStore extends NavbarQueryState {}

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
export const useGlobalStore = create<GlobalStore>((set, get, api) => ({
    ...createNavbarQueryStateStore(set, get, api),
}))

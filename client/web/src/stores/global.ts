import create from 'zustand'

import { NavbarQueryState, createNavbarQueryStateStore } from './navbarSearchQueryState'

interface GlobalStore extends NavbarQueryState {}

export const useGlobalStore = create<GlobalStore>((set, get) => ({
    ...createNavbarQueryStateStore(set, get),
}))

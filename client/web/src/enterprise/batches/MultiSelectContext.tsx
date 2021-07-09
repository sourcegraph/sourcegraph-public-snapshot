import React from 'react'

export type MultiSelectContextSelected = Set<string> | 'all'

export interface MultiSelectContextState {
    selected: MultiSelectContextSelected
    visible: Set<string>
    onDeselectAll: () => void
    onDeselectVisible: () => void
    onDeselect: (id: string) => void
    onSelectAll: () => void
    onSelectVisible: () => void
    onSelect: (id: string) => void
    onLoad: (ids: string[]) => void
}

export const MultiSelectContext = React.createContext<MultiSelectContextState>({
    selected: new Set(),
    visible: new Set(),
    onDeselectAll: () => {},
    onDeselectVisible: () => {},
    onDeselect: (id: string) => {},
    onSelectAll: () => {},
    onSelectVisible: () => {},
    onSelect: (id: string) => {},
    onLoad: (ids: string[]) => {},
})

export const useMultiSelectProvider = () => {}

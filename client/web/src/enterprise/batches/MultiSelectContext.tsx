import React from 'react'

export type MultiSelectContextSelected = Set<string> | 'all'

export interface MultiSelectContextState {
    selected: MultiSelectContextSelected
    onDeselectAll: () => void
    onDeselect: (id: string) => void
    onSelectAll: () => void
    onSelect: (id: string) => void
}

export const MultiSelectContext = React.createContext<MultiSelectContextState>({
    selected: new Set(),
    onDeselectAll: () => {},
    onDeselect: (id: string) => {},
    onSelectAll: () => {},
    onSelect: (id: string) => {},
})

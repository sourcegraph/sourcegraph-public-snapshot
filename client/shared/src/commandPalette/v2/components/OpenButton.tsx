import React from 'react'
import ConsoleIcon from 'mdi-react/ConsoleIcon'
import { useCommandPaletteStore } from '../store'

export const OpenCommandPaletteButton: React.FC = () => {
    const toggleCommandPaletteIsOpen = useCommandPaletteStore(state => state.toggleIsOpen)
    return (
        <button type="button" className="btn btn-link p-1" onClick={toggleCommandPaletteIsOpen as () => void}>
            <ConsoleIcon className="icon-inline-md" />
        </button>
    )
}

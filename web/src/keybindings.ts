import { Key, ModifierKey } from '@slimsag/react-shortcuts'

interface Keybinding {
    held?: ModifierKey[]
    ordered: Key[]
}

type Action = 'commandPalette' | 'switchTheme'

/**
 * Global keybinding actions. This is a map of action names to the keys that trigger them. The
 * actions are their own namespace for now, but it will be merged with extension actions/commands in
 * the future.
 */
export const keybindings: Record<Action, Keybinding[]> = {
    commandPalette: [{ held: ['Control'], ordered: ['p'] }, { ordered: ['F1'] }, { held: ['Alt'], ordered: ['x'] }],
    switchTheme: [{ held: ['Alt'], ordered: ['t'] }],
}

/** A partial React props type for components that use or propagate keybindings. */
export interface KeybindingsProps {
    /** The global map of keybindings and their associated actions. */
    keybindings: Readonly<typeof keybindings>
}

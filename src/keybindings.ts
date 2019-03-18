import { Key, ModifierKey } from '@shopify/react-shortcuts'

/**
 * A map of action names to the keys that trigger them. The actions are their own namespace for now, but it will be
 * merged with extension actions/commands in the future.
 */
export interface Keybindings extends Record<'commandPalette', Keybinding[]> {}

interface Keybinding {
    held?: ModifierKey[]
    ordered: Key[]
}

/**
 * Global keybinding actions.
 */
export const keybindings: Keybindings = {
    commandPalette: [{ held: ['Control'], ordered: ['p'] }, { ordered: ['F1'] }, { held: ['Alt'], ordered: ['x'] }],
}

/** A partial React props type for components that use or propagate keybindings. */
export interface KeybindingsProps {
    /** The global map of keybindings and their associated actions. */
    keybindings: Readonly<Keybindings>
}

import hotkeys, { type HotkeysEvent, type KeyHandler } from 'hotkeys-js'
import { onDestroy } from 'svelte'

import { isLinuxPlatform, isMacPlatform, isWindowsPlatform } from './common'

const LINUX_KEYNAME_MAP: Record<string, string> = {
    ctrl: 'Ctrl',
    shift: 'Shift',
    alt: 'Alt',
}
const WINDOWS_KEYNAME_MAP: Record<string, string> = LINUX_KEYNAME_MAP
const MAC_KEYNAME_MAP: Record<string, string> = {
    ctrl: '⌃',
    shift: '⇧',
    alt: '⌥',
    cmd: '⌘',
}

/**
 * Formats a key combination for display, properly replacing the key names with their platform-specific
 * counterparts.
 *
 * @param keys The key combination to format.
 * @returns A platform-specific string representing the key combination.
 */
export function formatShortcut(keys: Keys): string {
    return formatShortcutParts(keys).join(isMacPlatform() ? '' : '+')
}

/**
 * Formats a key combination for display, properly replacing the key names with their platform-specific
 * counterparts. Returns an array of strings, each representing a part of the key combination.
 *
 * @param keys The key combination to format.
 * @returns An array of strings, each representing a part of the key combination.
 */
export function formatShortcutParts(keys: Keys): string[] {
    const key = evaluateKey(keys)

    const parts = key.split('+')
    const out: string[] = []

    const keymap = isMacPlatform() ? MAC_KEYNAME_MAP : isLinuxPlatform() ? LINUX_KEYNAME_MAP : WINDOWS_KEYNAME_MAP

    for (const part of parts) {
        const lower = part.toLowerCase()
        if (keymap[lower]) {
            out.push(keymap[lower])
        } else {
            out.push(part.toUpperCase())
        }
    }

    return out
}

export function evaluateKey(keys: { mac?: string; linux?: string; windows?: string; key: string }): string {
    if (keys.mac && isMacPlatform()) {
        return keys.mac
    }

    if (keys.linux && isLinuxPlatform()) {
        return keys.linux
    }

    if (keys.windows && isWindowsPlatform()) {
        return keys.windows
    }

    return keys.key
}

function isElement(t: any): t is Element {
    return t instanceof Element
}

/**
 * This is an internal function to check if an Element has attributes that indicate it is a content field.
 * It's exported for testing only.
 * @param target Element
 */
function isContentElement(target: Element): boolean {
    return (
        target.getAttribute('contenteditable') === 'true' ||
        // textarea and input are from the HTML standard, textbox is from svelte
        ['textarea', 'input', 'textbox'].includes(target.getAttribute('role') ?? '') ||
        ['INPUT', 'TEXTAREA'].includes(target.tagName)
    )
}

/**
 * This function determines if the field that's focussed by the KeyboardEvent is some kind of input.
 * The implementation makes some assumptions about how the UI sets up content fields, which are also
 * specific to Svelte. It may need adjustment in the future.
 * @param event KeyboardEvent
 */
function isContentField(event: KeyboardEvent): boolean {
    if (!event?.target) {
        return false
    }
    const target = event.target
    if (isElement(target)) {
        return isContentElement(target)
    }
    return false
}

function wrapHandler(handler: KeyHandler, allowDefault: boolean = false, ignoreInputFields: boolean = true) {
    return (keyboardEvent: KeyboardEvent, hotkeysEvent: HotkeysEvent) => {
        // "Pass through" ignored events to allow them being processed by the target element
        if (ignoreInputFields && isContentField(keyboardEvent)) {
            return true
        }

        // When we use hotkeys.trigger, the event is null. That's why we need to check if the event and its function exist.
        if (!allowDefault && keyboardEvent?.preventDefault) {
            // Prevent the default refresh event under WINDOWS system
            keyboardEvent.preventDefault()
        }

        return handler(keyboardEvent, hotkeysEvent) ?? allowDefault
    }
}

export interface Keys {
    /**
     * The default key which should trigger the action.
     */
    key: string
    /**
     * An override for Mac users. The OS is resolved via https://developer.mozilla.org/en-US/docs/web/api/navigator/platform.
     */
    mac?: string
    /**
     * An override for Linux users. The OS is resolved via https://developer.mozilla.org/en-US/docs/web/api/navigator/platform.
     */
    linux?: string
    /**
     * An override for Windows users. The OS is resolved via https://developer.mozilla.org/en-US/docs/web/api/navigator/platform.
     */
    windows?: string
}

interface HotkeyOptions {
    /**
     * The keys that should trigger the handler.
     */
    keys: Keys
    /**
     * The action that should be triggered when the keys are pressed.
     */
    handler: KeyHandler
}

interface HotkeySetupOptions extends HotkeyOptions {
    /**
     * Whether the default browser behavior should execute.
     */
    allowDefault?: boolean
    /**
     * Whether the handler should be executed when the user focuses an input field.
     */
    ignoreInputFields?: boolean
}

/**
 * Creates a global keyboard shortcut. Needs to be called during
 * component initialization.
 */
export function registerHotkey({ keys, handler, allowDefault, ignoreInputFields }: HotkeySetupOptions): {
    bind: (options: HotkeyOptions) => void
    unregister: () => void
} {
    let currentKey = evaluateKey(keys)
    if (
        hotkeys
            .getAllKeyCodes()
            .map(k => k.shortcut)
            .includes(currentKey)
    ) {
        // Instead of printing an error, we can also use hotkey's "single" option, which will automatically unregister any
        // existing hotkey with the same key and scope.
        console.error(`The hotkey "${currentKey}" has already been registered by another Hotkey component.`)
    }
    let wrappedHandler = wrapHandler(handler, allowDefault, ignoreInputFields)

    // By default, hotkeys-js ignores input fields. Unfortunately this filter can only be set globally, and will apply to all hotkeys.
    // We work around this by always checking input fields, and then applying a custom filter in the wrappedHandler function.
    hotkeys.filter = function (_) {
        return true
    }

    onDestroy(() => {
        if (currentKey && wrappedHandler) {
            hotkeys.unbind(currentKey, wrappedHandler)
        }
    })

    if (currentKey) {
        hotkeys(currentKey, wrappedHandler)
    }

    return {
        /**
         * Use this function to change the shortcut and handler of a function. A use case for this may be when
         * a user changes their hotkey maps.
         */
        bind({ keys: bindKeys, handler: bindHandler }: HotkeyOptions) {
            if (currentKey) {
                hotkeys.unbind(currentKey, wrappedHandler)
            }
            currentKey = evaluateKey(bindKeys)
            wrappedHandler = wrapHandler(bindHandler, allowDefault, ignoreInputFields)
            hotkeys(currentKey, wrappedHandler)
        },
        /**
         * Use this function if you want to dynamically unregister a hotkey. You don't have to clean up after yourself:
         * The hotkey will be automatically removed when the lifecycle of a component ends (`onDestroy` hook).
         */
        unregister() {
            if (currentKey) {
                hotkeys.unbind(currentKey, wrappedHandler)
            }
        },
    }
}

export const exportedForTesting = {
    isContentElement,
    wrapHandler,
}

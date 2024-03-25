import { onDestroy } from 'svelte'
import hotkeys, { type HotkeysEvent, type KeyHandler } from 'hotkeys-js'
import { isLinuxPlatform, isMacPlatform, isWindowsPlatform } from '$lib/common'

function evaluateKey(keys: { mac?: string; linux?: string; windows?: string; key?: string }): string {
    if (keys.mac && isMacPlatform()) {
        return keys.mac
    }

    if (keys.linux && isLinuxPlatform()) {
        return keys.linux
    }

    if (keys.windows && isWindowsPlatform()) {
        return keys.windows
    }

    return keys.key ?? ''
}

function isContentField(event: KeyboardEvent): boolean {
    const target = event.target
    if (!target) {
        return false
    }
    if (target instanceof Element) {
        return (
            target.getAttribute('contenteditable') === 'true' ||
            // textarea and input are from the HTML standard, textbox is from svelte
            ['textarea', 'input', 'textbox'].includes(target.getAttribute('role') ?? '')
        )
    }
    return false
}

function wrapHandler(handler: KeyHandler, allowDefault: boolean = false, ignoreInputFields: boolean = true) {
    return (keyboardEvent: KeyboardEvent, hotkeysEvent: HotkeysEvent) => {
        if (!allowDefault) {
            // Prevent the default refresh event under WINDOWS system
            keyboardEvent.preventDefault()
        }

        if (!(ignoreInputFields && isContentField(keyboardEvent))) {
            handler(keyboardEvent, hotkeysEvent)
        }

        // Returning false stops the event and prevents default browser events on macOS.
        // It doesn't work for all default though, e.g. command+t will still open a new tab.
        return allowDefault
    }
}

interface Keys {
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
        bind({ keys: bindKeys, handler: bindHandler }: HotkeyOptions) {
            if (currentKey) {
                hotkeys.unbind(currentKey, wrappedHandler)
            }
            currentKey = evaluateKey(bindKeys)
            wrappedHandler = wrapHandler(bindHandler, allowDefault, ignoreInputFields)
            hotkeys(currentKey, wrappedHandler)
        },
    }
}

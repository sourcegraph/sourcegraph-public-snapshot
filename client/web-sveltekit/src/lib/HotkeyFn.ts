import {onDestroy} from 'svelte';
import hotkeys, {type HotkeysEvent, type KeyHandler} from 'hotkeys-js';
import {isLinuxPlatform, isMacPlatform, isWindowsPlatform} from '$lib/common';

function evaluateKey(keys: { mac?: string, linux?: string, windows?: string, key?: string }): string {
    if (keys.mac && isMacPlatform()) {
        return keys.mac;
    }

    if (keys.linux && isLinuxPlatform()) {
        return keys.linux;
    }

    if (keys.windows && isWindowsPlatform()) {
        return keys.windows;
    }

    return keys.key ?? '';
}

function isContentField(event: KeyboardEvent): boolean {
    const target = event.target;
    if (!target) {
        return false;
    }
    // todo: getAttribute doesn't seem to be supported by our typings, but works well in the browser. Figure out how to fix this.
    return target.getAttribute('contenteditable')
        // todo: figure out all roles that are editable (is this svelte specific?)
        || ['textarea', 'input', 'textbox'].includes(target.getAttribute('role'));
}

interface Keys {
    key: string,
    mac?: string,
    linux?: string,
    windows?: string,
}

function wrapHandler(handler: KeyHandler, allowDefault: boolean = false, ignoreInputFields: boolean = true) {
    return (keyboardEvent: KeyboardEvent, hotkeysEvent: HotkeysEvent) => {
        if (!allowDefault) {
            // Prevent the default refresh event under WINDOWS system
            keyboardEvent.preventDefault();
        }

        if (!(ignoreInputFields && isContentField(keyboardEvent))) {
            handler(keyboardEvent, hotkeysEvent);
        }

        // Returning false stops the event and prevents default browser events on macOS.
        // It doesn't work for all default though, e.g. command+t will still open a new tab.
        return allowDefault;
    }
}

interface HotkeyOptions {
    keys: Keys,
    handler: KeyHandler,
    allowDefault?: boolean,
    ignoreInputFields?: boolean
}

interface HotkeyBindOptions {
    keys: Keys,
    handler: KeyHandler,
}

/**
 * Creates a global keyboard shortcut. Needs to be called during
 * component initialization.
 */
export function registerHotkey({keys, handler, allowDefault, ignoreInputFields}: HotkeyOptions): {
    bind: (options: HotkeyBindOptions) => void
} {
    let currentKey = evaluateKey(keys);
    let wrappedHandler = wrapHandler(handler, allowDefault, ignoreInputFields);

    // By default, hotkeys-js ignores input fields. Unfortunately this filter can only be set globally, and will apply to all hotkeys.
    // We work around this by always checking input fields, and then applying a custom filter in the wrappedHandler function.
    hotkeys.filter = function (_) {
        return true;
    }

    onDestroy(() => {
        if (currentKey && wrappedHandler) {
            hotkeys.unbind(currentKey, wrappedHandler);
        }
    })

    if (currentKey) {
        hotkeys(currentKey, wrappedHandler);
    }

    return {
        bind({keys: bindKeys, handler: bindHandler}: HotkeyBindOptions) {
            if (currentKey) {
                hotkeys.unbind(currentKey, wrappedHandler);
            }
            currentKey = evaluateKey(bindKeys);
            wrappedHandler = wrapHandler(bindHandler, allowDefault, ignoreInputFields);
            hotkeys(currentKey, wrappedHandler);
        },
    }
}

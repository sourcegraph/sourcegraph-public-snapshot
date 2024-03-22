import {onDestroy} from 'svelte';
import hotkeys, {type HotkeysEvent, type KeyHandler} from 'hotkeys-js';
import {isLinuxPlatform, isMacPlatform, isWindowsPlatform} from '$root/client/common';

export function evaluateKey(keys: { mac?: string, linux?: string, windows?: string, key?: string }): string {
    if (isMacPlatform() && keys.mac) {
        return keys.mac;
    }

    if (isLinuxPlatform() && keys.linux) {
        return keys.linux;
    }

    if (isWindowsPlatform() && keys.windows) {
        return keys.windows;
    }

    return keys.key ?? '';
}

/**
 * Creates a global keyboard shortcut. Needs to be called during
 * component initialization.
 */
export function createHotkey(currentKey: string, currentHandler: KeyHandler, preventDefault: boolean = true, ignoreInputFields: boolean = true) {
    onDestroy(() => {
        if (currentKey) {
            hotkeys.unbind(currentKey, currentHandler)
        }
    })

    const run = currentHandler;
    currentHandler = (event: KeyboardEvent, handler: HotkeysEvent) => {
        if (preventDefault) {
            event.preventDefault();
        }

        if (!ignoreInputFields) {
            console.log('this should have triggered in an input field')
            // todo: implement filtering with an early return
        }

        run(event, handler);
    }

    if (currentKey && currentHandler) {
        hotkeys(currentKey, currentHandler)
    }

    return {
        bind(key: string, handler: KeyHandler) {
            if (currentKey) {
                hotkeys.unbind(currentKey, currentHandler)
            }
            hotkeys(key, handler)
            currentKey = key
            currentHandler = handler
        },
    }
}

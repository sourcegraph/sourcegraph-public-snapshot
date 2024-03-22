<script lang="ts">
    import hotkeys from 'hotkeys-js';
    import type {KeyHandler, HotkeysEvent} from 'hotkeys-js';
    import {onDestroy, onMount} from 'svelte';
    import {isLinuxPlatform, isMacPlatform, isWindowsPlatform} from '$root/client/common';

    export let run: () => void;
    export let key: string = '';
    export let linux: string = '';
    export let mac: string = '';
    export let windows: string = '';

    export let unbindAfterInvoke: boolean = false;

    const evaluateKey: (keys: { mac: string, linux: string, windows: string, key: string }) => string = (keys) => {
        if (isMacPlatform() && keys.mac) {
            return keys.mac;
        } else if (isLinuxPlatform() && keys.linux) {
            return keys.linux;
        } else if (isWindowsPlatform() && keys.windows) {
            return keys.windows;
        } else {
            return keys.key;
        }
    }

    let evaluatedKey: string = evaluateKey({mac, linux, windows, key});
    $: evaluatedKey = evaluateKey({mac, linux, windows, key});
    export let preventDefault: boolean = true;

    const handler: KeyHandler = (event: KeyboardEvent, _handler: HotkeysEvent) => {
        if (preventDefault) {
            event.preventDefault();
        }

        run();

        if (unbindAfterInvoke) {
            hotkeys.unbind(evaluatedKey, handler)
        }
    }

    onMount(() => {
        if (hotkeys.getAllKeyCodes().map(k => k.shortcut).includes(evaluatedKey)) {
            console.error(`The hotkey "${evaluatedKey}" has already been registered by another Hotkey component.`);
        }

        hotkeys(evaluatedKey, handler);
    });
    onDestroy(() => hotkeys.unbind(evaluatedKey, handler));
</script>

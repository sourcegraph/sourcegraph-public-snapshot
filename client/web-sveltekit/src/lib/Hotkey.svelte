<script lang="ts">
    import hotkeys from 'hotkeys-js';
    import {onDestroy, onMount} from 'svelte';

    export let run: () => void;
    export let key: string;
    export let preventDefaultEffect: boolean = true;

    onMount(() => {
        const conflictingKey = hotkeys.getAllKeyCodes().find(code => key.includes(code.shortcut))

        if (conflictingKey) {
            alert(`The key ${key} conflicts with another already registered hotkey.`);
            return;
        }

        hotkeys(key, function(event, _handler){
            if (preventDefaultEffect) {
                event.preventDefault();
            }

            // we can check for modifiers and ignore hotkeys to avoid duplicate bindings
            if (hotkeys.command) {
                console.log('command is pressed!');
                return;
            }

            run();
        });
    })

    onDestroy(() => {
        hotkeys.unbind(key);
    });
</script>

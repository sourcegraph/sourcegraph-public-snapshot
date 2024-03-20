<script lang="ts">
    import hotkeys from 'hotkeys-js';
    import {onDestroy} from 'svelte';

    export let run: () => void;
    export let key: string;
    export let preventDefaultEffect: boolean = true;

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

    console.log('registered key', {key})

    onDestroy(() => {
        hotkeys.unbind(key);
    });
</script>

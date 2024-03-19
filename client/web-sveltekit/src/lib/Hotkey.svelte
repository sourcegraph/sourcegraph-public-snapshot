<script lang="ts">
    import hotkeys from 'hotkeys-js';
    import {onDestroy} from 'svelte';

    export let run: () => void;
    export let key: string;
    export let preventDefault: boolean = true;

    function getEvaluatedKey() {
        // Here we can look up custom mappings
        return key;
    }

    hotkeys(getEvaluatedKey(), function(event, _handler){
        if (preventDefault) {
            event.preventDefault();
        }
        run();
    });

    onDestroy(() => {
        hotkeys.unbind(key);
    });
</script>

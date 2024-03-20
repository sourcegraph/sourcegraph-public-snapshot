<script lang="ts">
    import Hotkey from '$lib/Hotkey.svelte';

    export let run: () => void;
    export let keyDefault: string;
    // is there a better way for optional props? currently I'm using a default value, and then make sure it gets ignored with the || oerator
    export let keyOverride: string = "";
    export let preventDefaultEffect: boolean = true;

</script>

{#key [keyDefault, keyOverride, run, preventDefaultEffect]}
    <Hotkey key={keyOverride || keyDefault} run={run} preventDefaultEffect={preventDefaultEffect}/>
{/key}

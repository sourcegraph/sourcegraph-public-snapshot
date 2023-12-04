<!--
@component
Decorator for forcing the story component to be destroyed an recreated
-->

<script lang="ts">
    import type { SvelteComponent } from 'svelte'

    export let component: typeof SvelteComponent<any>
</script>

{#key $$restProps}
    <svelte:component this={component} {...$$restProps} />
{/key}

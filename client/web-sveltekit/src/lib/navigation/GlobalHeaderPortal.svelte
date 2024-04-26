<script>
    import { onDestroy } from 'svelte'

    import { portal } from '$lib/dom'
    import { navigationExtensionElement, navigationModeStore, NavigationMode } from './GlobalHeader.svelte'

    navigationModeStore.set(NavigationMode.WithCustomContent)

    onDestroy(() => {
        navigationModeStore.set(NavigationMode.PlainNavigation)
    })
</script>

<div class="root" use:portal={{ container: $navigationExtensionElement }}>
    <slot />
</div>

<style>
    .root {
        display: contents;
    }
</style>

<script lang="ts">
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'

    type $$Props = SvelteHTMLElements['svg'] & {
        icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
        inline?: boolean
    }

    export let icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
    export let inline: boolean = false
</script>

<span><svelte:component this={icon} class="icon {inline ? 'icon-inline' : ''}" /></span>

<style lang="scss">
    span {
        $iconSize: var(--icon-size, 1.5rem);
        $iconInlineSize: var(--icon-inline-size, #{(16 / 14)}em);

        > :global(svg.icon) {
            width: $iconSize;
            height: $iconSize;

            color: var(--icon-fill-color, var(--color, inherit));
            fill: currentColor;
        }

        > :global(svg.icon-inline) {
            width: $iconInlineSize;
            height: $iconInlineSize;
            vertical-align: text-bottom;
        }
    }
</style>

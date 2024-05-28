<script lang="ts">
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'
    import style from './Icon2.module.scss'

    type $$Props = SvelteHTMLElements['svg'] & {
        icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
        inline?: boolean
    }

    export let icon: ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
    export let inline: boolean = false
</script>

<svelte:component this={icon} class="{style.icon} {inline ? style.iconInline : ''}" />

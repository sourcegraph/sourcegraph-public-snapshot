<!--
  @component
  Provides a convenient API to render SVG icons.

  Use Lucide icons by referring to them as `ILucide<IconName>`. For example:

      <Icon2 icon={ILucideChevronDown} />

  See https://lucide.dev/icons/ for a list of available icons.
-->
<script lang="ts" context="module">
    export type IconComponent = ComponentType<SvelteComponent<SvelteHTMLElements['svg'] & {[key: `data-${string}`]: any }>>
</script>

<script lang="ts">
    // todo(@fkling): Rename to Icon.svelte once all references to the old Icon.svelte have been removed.
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'

    import style from './Icon2.module.scss'

    type $$Props = SvelteHTMLElements['svg'] & {
        /**
         * The SVG icon component to render.
         */
        icon: IconComponent
        /**
         * Render the icon inline next to text.
         */
        inline?: boolean
    }

    export let icon: IconComponent
    export let inline: boolean = false
</script>

<svelte:component data-icon this={icon} class="{style.icon} {inline ? style.iconInline : ''}" {...$$restProps} />

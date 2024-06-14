<!--
  @component
  Provides a convenient API to render SVG icons.

  Use Lucide icons by referring to them as `ILucide<IconName>`. For example:

      <Icon icon={ILucideChevronDown} />

  See https://lucide.dev/icons/ for a list of available icons.
-->
<script lang="ts" context="module">
    export type IconComponent = ComponentType<
        SvelteComponent<SvelteHTMLElements['svg'] & { [key: `data-${string}`]: any }>
    >
</script>

<script lang="ts">
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'

    import style from './Icon.module.scss'

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

<svelte:component this={icon} data-icon class="{style.icon} {inline ? style.iconInline : ''}" {...$$restProps} />

<script lang="ts" context="module">
    import type { ComponentType, SvelteComponent } from 'svelte'
    import type { SvelteHTMLElements } from 'svelte/elements'

    export type IconPath = ComponentType<SvelteComponent<SvelteHTMLElements['svg']>>
</script>

<!--
  @component
  Provides a convenient API to render SVG icons.

  Use Lucide icons by referring to them as `ILucide<IconName>`. For example:

      <Icon2 icon={ILucideChevronDown} />
-->
<script lang="ts">
    // todo(@fkling): Rename to Icon.svelte once all references to the old Icon.svelte have been removed.

    import style from './Icon2.module.scss'

    type $$Props = SvelteHTMLElements['svg'] & {
        /**
         * The SVG icon component to render.
         */
        icon: IconPath
        /**
         * Render the icon inline next to text.
         */
        inline?: boolean
    }

    export let icon: IconPath
    export let inline: boolean = false
</script>

<svelte:component this={icon} class="{style.icon} {inline ? style.iconInline : ''}" />

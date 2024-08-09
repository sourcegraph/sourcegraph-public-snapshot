<script lang="ts">
    import type { SvelteHTMLElements } from 'svelte/elements'

    import type { ExternalServiceKind } from '$lib/graphql-types'
    import Icon from '$lib/Icon.svelte'

    import { getIconForServiceKind, getExternalServiceHumanName } from './index'

    type $$Props = SvelteHTMLElements['svg'] & {
        /**
         * The service kind to render an icon for
         */
        kind: ExternalServiceKind
        /**
         * Render the icon inline next to text.
         */
        inline?: boolean
    }

    export let kind: $$Props['kind']
    export let inline: $$Props['inline'] = false
</script>

<div>
    <Icon aria-label={getExternalServiceHumanName(kind)} icon={getIconForServiceKind(kind)} {inline} {...$$restProps} />
</div>

<style lang="scss">
    div {
        display: contents;
        --icon-color: currentColor;
    }
</style>

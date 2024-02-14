<script lang="ts">
    import { mdiFileCodeOutline } from '@mdi/js'

    import { getFileIconInfo } from '$lib/wildcard'

    import Icon from './Icon.svelte'

    export let language: string
    const icon = getFileIconInfo('', language)?.svg
    const svgPath = icon?.path ?? mdiFileCodeOutline
    const svgColor = icon?.color
</script>

<Icon {svgPath} --color={svgColor} {...$$restProps} />

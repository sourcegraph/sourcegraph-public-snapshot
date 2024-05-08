<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { getIconPathForCodeHost } from '$lib/repo/shared/codehost'

    export let repository: string
    export let disableTooltip: boolean = false

    $: hostName = repository.split('/')[0]
    $: svgPath = getIconPathForCodeHost(hostName)
</script>

{#if svgPath}
    <Tooltip tooltip={disableTooltip ? '' : hostName}><Icon aria-label={hostName} {svgPath} inline /></Tooltip>
{/if}

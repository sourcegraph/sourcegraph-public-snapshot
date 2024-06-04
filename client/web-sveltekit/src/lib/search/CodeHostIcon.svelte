<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { getIconPathForCodeHost } from '$lib/repo/shared/codehost'
    import Tooltip from '$lib/Tooltip.svelte'

    export let repository: string
    export let codeHost: string | undefined = undefined
    export let disableTooltip: boolean = false

    $: hostName = repository.split('/')[0]
    $: svgPath = getIconPathForCodeHost(codeHost ?? hostName)
</script>

{#if svgPath}
    <Tooltip tooltip={disableTooltip ? '' : hostName}><Icon aria-label={hostName} {svgPath} inline /></Tooltip>
{/if}

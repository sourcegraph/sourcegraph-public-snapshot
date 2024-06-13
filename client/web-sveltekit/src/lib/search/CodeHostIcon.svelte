<script lang="ts">
    import Icon2 from '$lib/Icon2.svelte'
    import { getIconForCodeHost } from '$lib/repo/shared/codehost'
    import Tooltip from '$lib/Tooltip.svelte'

    export let repository: string
    export let codeHost: string | undefined = undefined
    export let disableTooltip: boolean = false

    $: hostName = repository.split('/')[0]
</script>

<Tooltip tooltip={disableTooltip ? '' : hostName}>
    <Icon2 aria-label={hostName} icon={getIconForCodeHost(codeHost ?? hostName)} inline />
</Tooltip>

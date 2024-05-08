<script lang="ts">
    import { mdiStarOutline } from '@mdi/js'

    import { formatRepositoryStarCount } from '$lib/branded'
    import Icon from '$lib/Icon.svelte'

    export let repoStars: number
    export let small = false
</script>

<span>
    <Icon inline svgPath={mdiStarOutline} />
    {#if small}
        <small>&nbsp;{formatRepositoryStarCount(repoStars)}</small>
    {:else}
        &nbsp;{formatRepositoryStarCount(repoStars)}
    {/if}
</span>

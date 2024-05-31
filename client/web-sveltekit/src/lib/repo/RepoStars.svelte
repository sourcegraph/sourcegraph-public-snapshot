<script lang="ts">
    import { formatRepositoryStarCount } from '$lib/branded'
    import Icon2 from '$lib/Icon2.svelte'

    export let repoStars: number
    export let small = false
</script>

<span>
    <Icon2 icon={ILucideStar} inline />
    {#if small}
        <small>&nbsp;{formatRepositoryStarCount(repoStars)}</small>
    {:else}
        &nbsp;{formatRepositoryStarCount(repoStars)}
    {/if}
</span>

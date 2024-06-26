<script lang="ts">
    import { formatRepositoryStarCount } from '$lib/branded'
    import Icon from '$lib/Icon.svelte'

    export let repoStars: number
</script>

<div>
    <Icon inline icon={ILucideStar} aria-label="Repository stars" />
    {formatRepositoryStarCount(repoStars)}
</div>

<style lang="scss">
    div {
        display: flex;
        align-items: center;
        gap: 0.25em;
    }
</style>

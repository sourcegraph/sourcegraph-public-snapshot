<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { formatRepositoryStarCount } from '$lib/branded'
    import { mdiStar } from '@mdi/js'

    export let repoStars: number
</script>

<span>
    <Icon inline svgPath={mdiStar} --color="var(--yellow)" />&nbsp;{formatRepositoryStarCount(repoStars)}
</span>

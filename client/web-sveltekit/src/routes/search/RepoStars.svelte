<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { formatRepositoryStarCount } from '$lib/branded'
    import { mdiStarOutline } from '@mdi/js'

    export let repoStars: number
</script>

<span>
    <Icon inline svgPath={mdiStarOutline} />&nbsp;{formatRepositoryStarCount(repoStars)}
</span>

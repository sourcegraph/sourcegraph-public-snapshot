<script lang="ts">
    import { getContext, type ComponentType } from 'svelte'

    export let key: string
    export let priority: number
    export let component: ComponentType

    const repoActions = getContext('repo-actions')
    repoActions.setAction({ key, priority, component })
</script>

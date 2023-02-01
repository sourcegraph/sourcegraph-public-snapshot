<script lang="ts">
    import { getContext, type ComponentType } from 'svelte'
    import type { ActionStore } from './actions'

    export let key: string
    export let priority: number
    export let component: ComponentType

    const repoActions = getContext<ActionStore>('repo-actions')
    repoActions.setAction({ key, priority, component })
</script>

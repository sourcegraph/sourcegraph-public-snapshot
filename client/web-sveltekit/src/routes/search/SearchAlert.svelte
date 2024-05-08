<script lang="ts">
    import type { Alert as SearchAlert } from '$lib/shared'
    import { Alert } from '$lib/wildcard'
    import Markdown from '$lib/wildcard/Markdown.svelte'

    export let alert: SearchAlert
</script>

<Alert variant="warning">
    <div>
        <h3>{alert.title}</h3>
        <Markdown content={alert.description ?? ''} />
    </div>
</Alert>

<script lang="ts">
    import type { Editor } from '$lib/web'

    export let editor: Editor

    $: svg = `/images/editors/${editor.id}.svg`
</script>

{#if editor}
    <img class:icon-inline={true} height={24} width={24} src={svg} alt={editor.name} />
{/if}

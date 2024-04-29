<script lang="ts">
    import type { Editor } from '$lib/web'
    import DefaultEditorIcon from '$lib/repo/open-in-editor/DefaultEditorIcon.svelte'

    export let editorId: Editor['id']
</script>

{#await import(`./editors/${editorId}.svelte`) then { default: Icon }}
    <Icon />
{:catch}
    <DefaultEditorIcon />
{/await}

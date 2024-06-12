<script lang="ts">
    import FileIcon from './repo/FileIcon.svelte'

    export let language: string
</script>

<!--
    This component abstracts from FileIcon specifics when we only have the language available.
    This allows us to use the same fallback logic.
-->

<FileIcon file={{ name: '', languages: [language] }} />

<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import FileIcon from './repo/FileIcon.svelte'

    type $$Props = {
        language: string
    } & Omit<ComponentProps<FileIcon>, 'file'>

    export let language: string
</script>

<!--
    This component abstracts from FileIcon specifics when we only have the language available.
    This allows us to use the same fallback logic.
-->

<FileIcon file={{ name: '', languages: [language] }} {...$$restProps} />

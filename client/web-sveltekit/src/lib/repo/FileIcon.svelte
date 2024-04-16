<script lang="ts">
    import { getFileIconInfo } from '$lib/wildcard'
    import { FileIcon_GitBlob } from './FileIcon.gql'
    import Icon from '$lib/Icon.svelte'
    import { mdiFileCodeOutline } from '@mdi/js'

    export let file: FileIcon_GitBlob | null
    export let inline = false

    $: icon = (file && getFileIconInfo(file.name, file.languages.at(0) ?? '')?.svg) || {
        path: mdiFileCodeOutline,
        color: 'var(--gray-05)',
    }
</script>

<Icon svgPath={icon.path} {inline} aria-hidden --color={icon.color} />

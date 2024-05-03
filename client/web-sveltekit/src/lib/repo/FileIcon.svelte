<script lang="ts">
    import { mdiFileCodeOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import { getFileIconInfo } from '$lib/wildcard'

    import { FileIcon_GitBlob } from './FileIcon.gql'

    export let file: FileIcon_GitBlob | null
    export let inline = false
    export let size: string = '1rem'

    $: icon = (file && getFileIconInfo(file.name, file.languages.at(0) ?? '')?.svg) || {
        path: mdiFileCodeOutline,
        color: 'var(--gray-05)',
    }
</script>

<Icon svgPath={icon.path} {inline} aria-hidden --color={icon.color} --icon-size={size} />

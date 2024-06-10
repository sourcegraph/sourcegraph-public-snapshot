<script lang="ts">
    import { getFileIconInfo, DEFAULT_ICON_COLOR } from '$lib/wildcard'
    import { FileIcon_GitBlob } from './FileIcon.gql'
    import Icon2 from '$lib/Icon2.svelte'

    export let file: FileIcon_GitBlob | null
    export let inline = false

    $: icon = (file && getFileIconInfo(file.name, file.languages.at(0) ?? '')) ?? {
        icon: ILucideFileCode,
        color: DEFAULT_ICON_COLOR,
    }
</script>

<Icon2 icon={icon.icon} {inline} aria-hidden style="color: {icon.color}" />

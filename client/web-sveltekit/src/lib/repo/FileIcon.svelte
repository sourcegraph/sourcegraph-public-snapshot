<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import Icon2 from '$lib/Icon2.svelte'
    import { getFileIconInfo, DEFAULT_ICON_COLOR } from '$lib/wildcard'

    import { FileIcon_GitBlob } from './FileIcon.gql'

    type $$Props = {
        file: FileIcon_GitBlob | null
    } & Omit<ComponentProps<Icon2>, 'icon'>

    export let file: FileIcon_GitBlob | null

    $: icon = (file && getFileIconInfo(file.name, file.languages.at(0) ?? '')) ?? {
        icon: ILucideFileCode,
        color: DEFAULT_ICON_COLOR,
    }
</script>

<Icon2 icon={icon.icon} aria-hidden style="color: {icon.color}" {...$$restProps} />

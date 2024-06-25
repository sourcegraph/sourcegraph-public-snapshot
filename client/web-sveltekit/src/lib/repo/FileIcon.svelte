<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import Icon from '$lib/Icon.svelte'
    import { getFileIconInfo, DEFAULT_ICON_COLOR } from '$lib/wildcard'

    import { FileIcon_GitBlob } from './FileIcon.gql'

    type $$Props = {
        file: FileIcon_GitBlob | null
    } & Omit<ComponentProps<Icon>, 'icon'>

    export let file: FileIcon_GitBlob | null

    $: icon = (file && getFileIconInfo(file.name, file.languages.at(0) ?? '')) ?? {
        icon: ILucideFileCode,
        color: DEFAULT_ICON_COLOR,
    }
</script>

<Icon icon={icon.icon} aria-hidden style="color: var(--file-icon-color, {icon.color})" {...$$restProps} />

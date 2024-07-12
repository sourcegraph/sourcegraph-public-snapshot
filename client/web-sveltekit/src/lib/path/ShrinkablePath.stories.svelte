<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import { sizeToFit } from '$lib/dom'
    import Icon from '$lib/Icon.svelte'

    import { pathHrefFactory } from '.'
    import ShrinkablePath from './ShrinkablePath.svelte'

    export const meta = {
        component: ShrinkablePath,
    }

    const path = 'very/very/very/long/displayed/path'

    let shrinkableDefault: ShrinkablePath
    let shrinkableLinkified: ShrinkablePath
</script>

<Story name="Default">
    <div use:sizeToFit={{ grow: () => shrinkableDefault.grow(), shrink: () => shrinkableDefault.shrink() }}>
        <ShrinkablePath bind:this={shrinkableDefault} {path} />
    </div>
</Story>

<Story name="Linkified">
    <div use:sizeToFit={{ grow: () => shrinkableLinkified.grow(), shrink: () => shrinkableLinkified.shrink() }}>
        <ShrinkablePath
            bind:this={shrinkableLinkified}
            {path}
            pathHref={pathHrefFactory({
                repoName: 'myrepo',
                revision: 'main',
                fullPath: path,
                fullPathType: 'blob',
            })}
        >
            <Icon icon={ILucideEye} inline />
        </ShrinkablePath>
    </div>
</Story>

<style lang="scss">
    div {
        width: 200px;
        overflow: hidden;
        resize: horizontal;
    }
</style>

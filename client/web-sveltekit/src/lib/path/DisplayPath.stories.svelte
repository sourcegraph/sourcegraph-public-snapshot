<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import Icon from '$lib/Icon.svelte'

    import { pathHrefFactory } from '.'
    import DisplayPath from './DisplayPath.svelte'

    export const meta = {
        component: DisplayPath,
    }
</script>

<Story name="Default">
    <DisplayPath path="my/displayed/path" />
</Story>

<Story name="Copy button">
    <DisplayPath path="my/displayed/path" showCopyButton />
</Story>

<Story name="File Icon">
    <DisplayPath path="my/displayed/path">
        <Icon slot="file-icon" icon={ILucideEye} inline />
    </DisplayPath>
</Story>

<Story name="Linkified">
    <DisplayPath
        path="my/displayed/path"
        pathHref={pathHrefFactory({
            repoName: 'myrepo',
            revision: 'main',
            fullPath: 'my/displayed/path',
            fullPathType: 'blob',
        })}
    >
        <Icon icon={ILucideEye} inline />
    </DisplayPath>
</Story>

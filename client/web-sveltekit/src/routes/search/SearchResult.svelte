<script lang="ts">
    import { mdiBitbucket, mdiGithub, mdiGitlab, mdiStar } from '@mdi/js'

    import type { SearchMatch } from '$lib/shared'
    import Icon from '$lib/Icon.svelte'
    import { formatRepositoryStarCount } from '$lib/branded'
    import Tooltip from '$lib/Tooltip.svelte'

    function codeHostIcon(repoName: string): { hostName: string; svgPath?: string } {
        const hostName = repoName.split('/')[0]
        const iconMap: { [key: string]: string } = {
            'github.com': mdiGithub,
            'gitlab.com': mdiGitlab,
            'bitbucket.org': mdiBitbucket,
        }
        return { hostName, svgPath: iconMap[hostName] }
    }

    export let result: SearchMatch

    $: icon = codeHostIcon(result.repository)
</script>

<article>
    <div class="header">
        {#if icon.svgPath}
            <Tooltip tooltip={icon.hostName}>
                <Icon class="text-muted" aria-label={icon.hostName} svgPath={icon.svgPath} inline />{' '}
            </Tooltip>
        {/if}
        <div class="title">
            <slot name="title" />
            {#if result.repoStars}
                <div class="star">
                    <Icon inline svgPath={mdiStar} --color="var(--yellow)" />
                    {formatRepositoryStarCount(result.repoStars)}
                </div>
            {/if}
        </div>
    </div>
    <slot />
</article>

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        padding: 0.5rem 0.5rem 0.5rem 0;
        position: sticky;
        top: 0;
        background-color: var(--body-bg);
    }

    .title {
        flex: 1 1 auto;
        overflow: hidden;
        display: flex;
        flex-wrap: wrap;

        // .title-inner
        overflow-wrap: anywhere;

        // .muted-repo-file-link
        color: var(--text-muted);

        :global(a) {
            color: var(--text-muted);

            &:hover {
                color: var(--text-muted);
            }
        }
    }

    .star {
        margin-left: auto;
    }
</style>

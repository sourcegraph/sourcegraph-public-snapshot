<script lang="ts">
    import { mdiFileDocumentOutline } from '@mdi/js'

    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import FileTable from '$lib/repo/ui/FileTable.svelte'
    import { getRelativeTime } from '$lib/relativeTime.js'
    import UserAvatar from '$lib/UserAvatar.svelte'

    export let data

    $: commitWithTree = data.deferred.commitWithTree.then(result => isErrorLike(result) ? null : result)
</script>

<section>
    <div class="main">
        {#await commitWithTree then commit}
            {#if commit}
                <div class="card">
                <h3 class="header">Latest Commit</h3>
                <div class="content commit-summary">
                    <span><UserAvatar user={commit.author.person} /></span>
                    <p><a href={commit.url}>{commit.subject}</a></p>
                    <a href={commit.url}>{commit.abbreviatedOID}</a>
    &nbsp;&middot;&nbsp;
                    <span>{getRelativeTime(new Date(commit.author.date))}</span>
                </div>
                </div>
                <section class="mb-3">
                    <FileTable treeOrError={commit.tree}/>
                </section>

                {/if}
            {/await}
        {#await data.deferred.readmeBlob then blob}
            {#if blob}
                <div class="readme card">
                    <h3 class="header">
                        <Icon svgPath={mdiFileDocumentOutline} inline />
                        {blob.name}
                    </h3>
                    <div class="content">
                        {#if blob?.richHTML}
                            {@html blob.richHTML}
                        {:else if blob.content}
                            <pre>{blob.content}</pre>
                        {/if}
                    </div>
                </div>
            {/if}
        {/await}
    </div>
    <div class="side">
        {#if data.resolvedRevision}
            <div class="side-card">
                <h3>Description</h3>
                <p>
                    {data.resolvedRevision.repo.description}
                </p>
            </div>
        {/if}

    </div>
</section>

<style lang="scss">
    section {
        display: flex;
        overflow: hidden;
    }

    .main {
        flex: 1;
        flex-basis: 0px;
        overflow-y: auto;
    }

    .side {
        flex-shrink: 0;
        width: 200px;
        margin: 0 1rem;

        .side-card {
        }
    }

    .commit-summary {
        display: flex;
        margin-bottom: 1rem;
        align-items: center;
        padding: 0.5rem;
        border: 1px solid var(--border-color);
        border-radius: var(--border-radius);

        p {
            margin: 0;
            margin-left: 0.5rem;
            flex: 1;
        }
    }
    .card {
        .header {
            border: 1px solid var(--border-color);
            background-color: var(--code-bg);
            position: sticky;
            top: 0;
            padding: 0.5rem;
            border-bottom: 1px solid var(--border-color);

        }

        h3.header {
            margin: 0;
        }

        .content {
            border: 1px solid var(--border-color);
            border-top: none;
            background-color: var(--code-bg);
            padding: 0.5rem;

        }
    }

    ul.commits {
        padding: 0;
        margin: 0;
        list-style: none;

        li {
            border-bottom: 1px solid var(--border-color);
            padding: 0.5rem 0;

            &:last-child {
                border: none;
            }
        }
    }

    ul.files {
        padding: 0;
        margin: 0;
        list-style: none;
        columns: 3;
    }
</style>

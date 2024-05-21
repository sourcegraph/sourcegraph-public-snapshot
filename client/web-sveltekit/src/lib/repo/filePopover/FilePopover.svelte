<script lang="ts" context="module">
    import { type DirPopoverFragment, type FilePopoverFragment, FileOrDirPopoverQuery } from './FilePopover.gql'

    export interface FilePopoverInputs {
        repoName: string
        revision: string
        filePath: string
    }

    export async function fetchPopoverData(args: FilePopoverInputs): Promise<DirPopoverFragment | FilePopoverFragment> {
        const client = getGraphQLClient()
        const result = await client.query(FileOrDirPopoverQuery, args)
        if (result.error) {
            throw new Error('could not fetch file or dir popover', result.error)
        }
        const fragment = result.data?.repository?.commit?.path
        if (!fragment) {
            throw new Error('entry does not exist')
        }
        return fragment
    }
</script>

<script lang="ts">
    import { mdiFolder } from '@mdi/js'

    import Avatar from '$lib/Avatar.svelte'
    import { pluralize } from '$lib/common'
    import { getGraphQLClient } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import { displayRepoName } from '$lib/shared'
    import Timestamp from '$lib/Timestamp.svelte'
    import { formatBytes } from '$lib/utils'
    import Badge from '$lib/wildcard/Badge.svelte'

    import FileIcon from '../FileIcon.svelte'

    import NodeLine from './NodeLine.svelte'

    export let repoName: string
    export let filePath: string
    export let entry: FilePopoverFragment | DirPopoverFragment

    function splitPath(filePath: string): [string, string] {
        let parts = filePath.split('/')
        console.log({ parts })
        return [parts.slice(0, parts.length - 1).join('/'), parts[parts.length - 1]]
    }

    $: [dirName, baseName] = splitPath(filePath)
    $: lastCommit = entry.history.nodes[0].commit
</script>

<div class="root section muted">
    <div class="repo-and-path section mono">
        <small>
            {displayRepoName(repoName).replaceAll('/', ' / ')}
            ·
            {dirName ? `${dirName.replaceAll('/', ' / ')}` : '/'}
        </small>
    </div>

    <div class="lang-and-file section">
        {#if entry.__typename === 'GitBlob'}
            <FileIcon file={entry} inline={false} --icon-size="1.5rem" />
            <div class="file mono">
                <div>{baseName}</div>
                <small>
                    {entry.languages[0] ? `${entry.languages[0]} ·` : ''}
                    {entry.totalLines}
                    {pluralize('Line', entry.totalLines)} ·
                    {formatBytes(entry.byteSize)}
                </small>
            </div>
        {:else if entry.__typename === 'GitTree'}
            <Icon svgPath={mdiFolder} --icon-fill-color="var(--primary)" --icon-size="1.5rem" />
            <div class="file mono">
                <div class="title">{baseName}</div>
                <small>
                    Subdirectories {entry.directories.length}
                    · Files {entry.files.length}
                </small>
            </div>
        {/if}
    </div>

    <div class="last-changed section">Last Changed @</div>

    <div class="commit">
        <div class="node-line"><NodeLine /></div>
        <div class="commit-info">
            <Badge variant="link">
                <a href={lastCommit.canonicalURL} target="_blank">
                    {lastCommit.abbreviatedOID}
                </a>
            </Badge>
            <div class="body">{lastCommit.subject}</div>
            <div class="author">
                <Avatar avatar={lastCommit.author.person} --avatar-size="1.0rem" />
                {lastCommit.author.person.displayName}
                ·
                <Timestamp date={lastCommit.author.date} />
            </div>
        </div>
    </div>
</div>

<style lang="scss">
    .root {
        width: 400px;
        background: var(--body-bg);

        .section {
            padding: 0.5rem 1rem;
        }

        .repo-and-path {
            border-bottom: 1px solid var(--border-color);
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .lang-and-file {
            display: flex;
            align-items: center;
            gap: 1rem;

            .file {
                display: flex;
                flex-direction: column;
                align-items: flex-start;
                gap: 0.25rem;

                div {
                    color: var(--text-body);
                }
            }
        }

        .last-changed {
            background-color: var(--secondary-4);
            border-bottom: 1px solid var(--border-color);
        }

        .commit {
            display: flex;
            align-items: stretch;
            justify-content: flex-start;

            .node-line {
                flex: 0 0 40px;
            }

            .commit-info {
                flex: 1;

                display: flex;
                flex-direction: column;
                align-items: flex-start;
                gap: 0.25rem;
                padding: 0.5rem 0.5rem 0.5rem 0;

                .author {
                    display: flex;
                    justify-content: flex-start;
                    align-items: center;
                    gap: 0.25rem;
                    font-size: var(--font-size-tiny);
                }
            }
        }
    }

    .mono {
        font-family: var(--monospace-font-family);
    }

    .title {
        color: var(--text-title);
    }

    .muted {
        color: var(--text-muted);
    }

    .body {
        color: var(--text-body);
    }
</style>

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
    import Avatar from '$lib/Avatar.svelte'
    import { pluralize } from '$lib/common'
    import { getGraphQLClient } from '$lib/graphql'
    import Icon from '$lib/Icon.svelte'
    import { pathHrefFactory } from '$lib/path'
    import DisplayPath from '$lib/path/DisplayPath.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { formatBytes } from '$lib/utils'
    import Badge from '$lib/wildcard/Badge.svelte'

    import FileIcon from '../FileIcon.svelte'

    import NodeLine from './NodeLine.svelte'

    export let repoName: string
    export let revision: string
    export let entry: FilePopoverFragment | DirPopoverFragment

    function splitPath(filePath: string): [string, string] {
        let parts = filePath.split('/')
        return [parts.slice(0, parts.length - 1).join('/'), parts[parts.length - 1]]
    }

    $: [dirName, baseName] = splitPath(entry.path)
    $: lastCommit = entry.history.nodes[0].commit
</script>

<div class="root">
    {#if dirName.length > 0}
        <div class="section">
            <DisplayPath
                path={dirName}
                pathHref={pathHrefFactory({ repoName, revision, fullPath: dirName, fullPathType: 'tree' })}
            />
        </div>
    {/if}

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
            <Icon icon={ILucideFolder} --icon-color="var(--primary)" --icon-size="1.5rem" />
            <div class="file mono">
                <div class="title">{baseName}</div>
                <small>
                    Subdirectories {entry.directories.length}
                    · Files {entry.files.length}
                </small>
            </div>
        {/if}
    </div>

    <div class="last-changed section">Last Changed</div>

    <div class="commit">
        <div class="node-line"><NodeLine /></div>
        <div class="commit-info">
            <Badge variant="link">
                <a href={lastCommit.canonicalURL} target="_blank">
                    {lastCommit.abbreviatedOID}
                </a>
            </Badge>
            <div class="body"><a href={lastCommit.canonicalURL}>{lastCommit.subject}</a></div>
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
        background: var(--color-bg-1);

        .section {
            padding: 0.375rem 1rem 0.5rem 1rem;
            border-bottom: 1px solid var(--border-color);
        }

        .lang-and-file {
            display: flex;
            align-items: center;
            gap: 0.75rem;

            .file {
                display: flex;
                flex-direction: column;
                align-items: flex-start;
                gap: 0.125rem;

                small {
                    color: var(--text-muted);
                }
            }
        }

        .last-changed {
            background-color: var(--secondary-4);
            color: var(--text-body);
            font-size: var(--font-size-xs);
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
                padding: 0.625rem 0.625rem 0.625rem 0;

                .author {
                    display: flex;
                    justify-content: flex-start;
                    align-items: center;
                    gap: 0.375rem;
                    font-size: var(--font-size-xs);
                    color: var(--text-muted);
                }
            }
        }
    }

    .mono {
        font-family: var(--monospace-font-family);
    }

    .body {
        color: var(--text-body);
        a {
            color: unset;
        }
    }
</style>

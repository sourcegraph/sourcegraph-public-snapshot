<script lang="ts">
    import Avatar from '$lib/Avatar.svelte'
    import { numberWithCommas } from '$lib/common'
    import type { GitRefType } from '$lib/graphql-types'
    import CodeHostIcon from '$lib/search/CodeHostIcon.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Badge } from '$lib/wildcard'
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    import type { GitReferencesTable_Ref } from './GitReferencesTable.gql'
    import { getHumanNameForCodeHost } from './shared/codehost'

    export let references: GitReferencesTable_Ref[]
    export let referenceType: GitRefType
    export let defaultBranch: string = ''

    const typeToDisplayName: Record<GitRefType, string> = {
        GIT_BRANCH: 'Branch',
        GIT_TAG: 'Tag',
        GIT_REF_OTHER: 'Reference',
    }

    function getMaxBehindAhead(references: GitReferencesTable_Ref[]): number {
        let max = 0
        for (const ref of references) {
            const behindAhead = ref.target.commit?.behindAhead
            if (behindAhead) {
                if (behindAhead.behind > max) {
                    max = behindAhead.behind
                }
                if (behindAhead.ahead > max) {
                    max = behindAhead.ahead
                }
            }
        }
        return max
    }

    $: showBehindAhead = references.some(ref => ref.target.commit?.behindAhead)
    $: max = getMaxBehindAhead(references)
</script>

<table>
    <thead>
        <tr>
            <th>{typeToDisplayName[referenceType]}</th>
            <th>Updated</th>
            {#if showBehindAhead}
                <th class="behindAhead">Behind/Ahead</th>
            {/if}
            <th><span class="visually-hidden">Actions</span></th>
        </tr>
    </thead>
    <tbody>
        {#each references as ref (ref.id)}
            {@const commit = ref.target.commit}
            {@const behind = commit?.behindAhead?.behind ?? 0}
            {@const ahead = commit?.behindAhead?.ahead ?? 0}

            <tr>
                <td class="revision">
                    <Badge variant="link">
                        <a href={ref.url} title={ref.displayName} slot="custom" let:class={className} class={className}>
                            {ref.displayName}
                        </a>
                    </Badge>
                    <CopyButton value={ref.displayName} label="Copy branch name" />
                </td>
                <td class="timestamp">
                    {#if commit}
                        <Avatar avatar={commit.author.person} />
                        <small><Timestamp date={commit.author.date} strict /></small>
                    {/if}
                </td>
                {#if showBehindAhead}
                    {@const isDefault = defaultBranch === ref.displayName}
                    <td class="behindAhead" class:default={isDefault}>
                        {#if ahead || behind}
                            <Tooltip tooltip={`Behind: ${behind} commits, Ahead: ${ahead} commits`}>
                                <div class="wrapper">
                                    <small class="behind">{numberWithCommas(behind)}</small>
                                    <small class="ahead">{numberWithCommas(ahead)}</small>
                                    <div class="bar behind" style:width="{(behind / max) * 100}%" />
                                    <div class="bar ahead" style:width="{(ahead / max) * 100}%" />
                                </div>
                            </Tooltip>
                        {:else if isDefault}
                            <Badge variant="secondary">default</Badge>
                        {/if}
                    </td>
                {/if}
                <td class="actions">
                    {#if commit}
                        <ul>
                            <li
                                ><Badge variant="link">
                                    <a href={commit.canonicalURL} title="View commit">{commit.abbreviatedOID}</a>
                                </Badge></li
                            >
                            {#each commit.externalURLs as { url, serviceKind }}
                                <li>
                                    <a href={url}>
                                        View on
                                        {#if serviceKind}
                                            <CodeHostIcon repository={serviceKind} disableTooltip />
                                            {getHumanNameForCodeHost(serviceKind)}
                                        {:else}
                                            code host
                                        {/if}
                                    </a>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                </td>
            </tr>
        {/each}
    </tbody>
</table>

<style lang="scss">
    table {
        max-width: 100%;
        border: 1px solid var(--border-color-2);
        background-color: var(--color-bg-1);

        display: grid;
        grid-template-columns: [revision] 1fr [timestamp] auto [behindAhead] auto [actions] auto;
        column-gap: 0.5rem;

        @media (--mobile) {
            grid-template-columns: auto auto;

            thead {
                display: none;
            }
        }
    }

    thead,
    tbody,
    tr {
        display: grid;
        grid-column: 1 / -1;
        grid-template-columns: subgrid;
    }

    thead {
        background-color: var(--color-bg-2);
    }

    tr {
        padding: 0.25rem 0.5rem;
        row-gap: 0.5rem;

        + tr {
            border-top: 1px solid var(--border-color-2);
        }

        @media (--sm-breakpoint-down) {
            padding: 0.5rem;
            column-gap: 0.5rem;
        }

        @media (--mobile) {
            grid-template-areas: 'revision revision' 'timestamp behindAhead' 'actions actions';
        }
    }

    th {
        font-weight: 500;
        font-size: var(--font-size-small);
        padding: 0.5rem;
    }

    th.behindAhead,
    .behindAhead.default {
        text-align: center;
    }

    td {
        align-self: center;

        &.revision {
            grid-area: revision;
        }
        &.timestamp {
            grid-area: timestamp;
        }
        &.behindAhead {
            grid-area: behindAhead;
        }
        &.actions {
            grid-area: actions;
        }

        @media (--mobile) {
            padding: 0;
        }
    }

    span {
        white-space: nowrap;
    }

    small {
        color: var(--text-muted);
        white-space: nowrap;

        &:hover {
            color: inherit;
        }
    }

    .revision {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        overflow: hidden;

        a {
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .behindAhead .wrapper {
        display: grid;
        grid-template-columns: 1fr 1fr;
        grid-template-rows: auto auto;
        gap: 0rem 0.5rem;
        justify-content: center;

        > * {
            flex: 1;
        }

        .behind {
            justify-self: end;
        }

        .bar {
            background-color: var(--color-bg-3);
            height: 0.5rem;
        }
    }

    .actions {
        ul {
            --icon-color: currentColor;

            font-size: var(--font-size-small);
            list-style: none;
            padding: 0;
            margin: 0;

            @media (--mobile) {
                display: flex;
                gap: 0.25rem;

                li:not(:first-child)::before {
                    content: '•';
                    padding-right: 0.25rem;
                    color: var(--text-muted);
                }
            }
        }
    }
</style>

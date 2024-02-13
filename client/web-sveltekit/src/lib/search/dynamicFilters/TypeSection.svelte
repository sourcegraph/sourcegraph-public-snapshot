<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import { resultTypeFilter } from '$lib/search/sidebar'

    import { type QueryStateStore, getQueryURL } from '../state'

    export let queryFromURL: string
    export let queryState: QueryStateStore
</script>

<ul>
    <!-- TODO: a11y -->
    {#each resultTypeFilter as filter}
        <li class:selected={filter.isSelected(queryFromURL)}>
            <a
                href={getQueryURL(
                    {
                        searchMode: $queryState.searchMode,
                        patternType: $queryState.patternType,
                        caseSensitive: $queryState.caseSensitive,
                        searchContext: $queryState.searchContext,
                        query: filter.getQuery($queryState.query),
                    },
                    true
                )}
            >
                <Icon svgPath={filter.icon} inline aria-hidden="true" />
                {filter.label}
            </a>
        </li>
    {/each}
</ul>

<style lang="scss">
    .selected {
        a {
            background-color: var(--primary);
            color: var(--primary-4);
            --color: var(--primary-4);
        }
    }

    ul {
        margin: 0;
        padding: 0;
        list-style: none;

        a {
            flex: 1;
            color: var(--sidebar-text-color);
            text-decoration: none;
            padding: 0.25rem 0.5rem;
            border-radius: var(--border-radius);
            // Controls icon color
            --color: var(--icon-color);

            &:hover {
                background-color: var(--secondary-4);
            }
        }

        li {
            display: flex;
            white-space: nowrap;

            &.selected {
                a {
                    background-color: var(--primary);
                    color: var(--primary-4);
                    --color: var(--primary-4);
                }
            }
        }
    }
</style>

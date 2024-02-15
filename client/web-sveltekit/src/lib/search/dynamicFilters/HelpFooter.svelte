<script>
    import LinkShareIcon from '$lib/icons/LinkShare.svelte'
    import OpenBookIcon from '$lib/icons/OpenBook.svelte'
</script>

<footer class="help">
    <a href="/help/code_search/reference/queries" class="link">
        <OpenBookIcon />
        <span>
            <h4>Need more advanced filters?</h4>
            <small>Explore the query syntax docs</small>
        </span>
        <LinkShareIcon class="icon" />
    </a>
</footer>

<style lang="scss">
    .help {
        &::before {
            content: '';
            height: 1px;
            background: var(--border-color);
            display: block;
        }

        &:hover {
            &::before {
                background: linear-gradient(91deg, var(--logo-blue) 3.11%, #c66fff 44.21%, #ff8578 83.64%);
            }

            .link {
                background: linear-gradient(
                    90deg,
                    rgba(0, 203, 236, 0.08) 3.11%,
                    rgba(198, 111, 255, 0.08) 44.21%,
                    rgba(255, 133, 120, 0.08) 83.64%
                );
            }

            --icon-color: var(--primary);
        }

        a {
            padding: 1rem;
            h4 {
                margin: 0;
            }

            display: flex;
            align-items: center;
            gap: 1rem;
            span {
                margin-right: auto;
            }
            :global(svg) {
                flex-shrink: 0;
            }

            text-decoration: none;
            color: var(--text-muted);
            font-size: 0.75rem;
        }
    }
</style>

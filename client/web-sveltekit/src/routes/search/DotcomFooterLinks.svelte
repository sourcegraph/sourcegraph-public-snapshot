<script lang="ts">
    const links = [
        {
            name: 'Docs',
            href: 'https://sourcegraph.com/docs',
            telemetryType: 1,
        },
        { name: 'About', href: 'https://sourcegraph.com', telemetryType: 2 },
        {
            name: 'Cody',
            href: 'https://sourcegraph.com/cody',
            telemetryType: 3,
        },
        {
            name: 'Enterprise',
            href: 'https://sourcegraph.com/get-started?t=enterprise',
            telemetryType: 4,
        },
        {
            name: 'Security',
            href: 'https://sourcegraph.com/security',
            telemetryType: 5,
        },
        {
            name: 'Discord',
            href: 'https://srcgr.ph/discord-server',
            telemetryType: 6,
        },
    ]

    function handleLinkClick(telemetryType: number): void {
        import('$lib/telemetry').then(({ TELEMETRY_RECORDER }) => {
            TELEMETRY_RECORDER.recordEvent('home.footer.CTA', 'click', { metadata: { type: telemetryType } })
        })
    }
</script>

{#if links.length > 0}
    <footer>
        {#each links as link}
            <a
                href={link.href}
                on:click={() => handleLinkClick(link.telemetryType)}
                rel="noopener noreferrer"
                target="_blank"
            >
                {link.name}
            </a>
        {/each}
    </footer>
{/if}

<style lang="scss">
    footer {
        display: flex;
        flex-direction: row;

        a {
            color: var(--text-muted);
            padding: 0 1rem;
            &:not(:last-child) {
                border-right: 1px solid var(--border-color);
            }
        }

        // In a small viewport, align links in a column and remove the separator
        @media (--xs-breakpoint-down) {
            flex-direction: column;
            gap: 0.5rem;
            align-items: center;

            a:not(:last-child) {
                border: none;
            }
        }
    }
</style>

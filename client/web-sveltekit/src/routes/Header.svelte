<script lang="ts">
    import { mdiBookOutline, mdiChartBar, mdiMagnify } from '@mdi/js'

    import {mark} from '$lib/images'
    import UserAvatar from '$lib/UserAvatar.svelte'
    import type { AuthenticatedUser } from '$lib/shared'
    import { PUBLIC_SG_ENTERPRISE } from '$env/static/public'

    import HeaderNavLink from './HeaderNavLink.svelte'

    export let authenticatedUser: AuthenticatedUser | null | undefined
</script>

<header>
    <a href="/search">
        <img src={mark} alt="Sourcegraph" width="25" height="25" />
    </a>
    <nav class="ml-2">
        <ul>
            <HeaderNavLink href="/search" svgIconPath={mdiMagnify}>Code search</HeaderNavLink>
            {#if PUBLIC_SG_ENTERPRISE}
                <!--
                    Example for conditionally showing navigation links.
                    Need to investigate whether a branch like this would be removed
                    in a production build due to dead-code elimination.
                -->
                <HeaderNavLink href="/contexts">Contexts</HeaderNavLink>
            {/if}
            <!--
            <li aria-current={$page.url.pathname === '/notebooks' ? 'page' : undefined}>
                <a href="/notebooks">
                    <svg aria-hidden="true" viewBox="0 0 25 25">
                        <path d={mdiBookOutline} />
                    </svg>
                    Notebooks
                </a>
            </li>
            <li aria-current={$page.url.pathname === '/code-monitoring' ? 'page' : undefined}>
                <a href="/code-monitoring">
                    <svg
                        width="20"
                        height="20"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <path
                            fill-rule="evenodd"
                            clip-rule="evenodd"
                            d="M18.01 8.01C18.01 8.29 18.23 8.51 18.51 8.51C18.79 8.51 19.01 8.29 19.01 8.01C19.01 4.15 15.87 1 12 1C11.72 1 11.5 1.22 11.5 1.5C11.5 1.78 11.72 2 12 2C15.31 2 18.01 4.7 18.01 8.01ZM16.1801 7.96002C15.9001 7.96002 15.6801 7.74002 15.6801 7.46002C15.6801 5.81002 14.3301 4.46002 12.6801 4.46002C12.4001 4.46002 12.1801 4.24002 12.1801 3.96002C12.1801 3.68002 12.4001 3.46002 12.6801 3.46002C14.8901 3.46002 16.6801 5.25002 16.6801 7.46002C16.6801 7.74002 16.4601 7.96002 16.1801 7.96002ZM4.83996 6.79999L13.34 15.3C12.39 15.88 11.29 16.18 10.15 16.18C8.49996 16.18 6.93996 15.54 5.76996 14.37C4.59996 13.2 3.94996 11.65 3.94996 9.98999C3.94996 8.84999 4.25996 7.74999 4.83996 6.79999ZM4.70996 4.54999C1.70996 7.54999 1.70996 12.43 4.70996 15.43C6.20996 16.93 8.17996 17.68 10.15 17.68C12.12 17.68 14.09 16.93 15.59 15.43L4.70996 4.54999ZM4 16.14C3.7 15.84 3.43 15.52 3.18 15.18L2.89 15.69L1 18.97H4.79H8.59L8.31 18.49C6.69 18.14 5.2 17.34 4 16.14ZM13.85 8.04999C13.85 9.01999 13.07 9.79999 12.1 9.79999C11.13 9.79999 10.35 9.01999 10.35 8.04999C10.35 7.07999 11.13 6.29999 12.1 6.29999C13.07 6.29999 13.85 7.07999 13.85 8.04999Z"
                        />
                    </svg>
                    Monitoring
                </a>
            </li>
            <li aria-current={$page.url.pathname === '/batch-changes' ? 'page' : undefined}>
                <a href="/batch-changes">
                    <svg
                        width="20"
                        height="20"
                        viewBox="0 -3 38 38"
                        fill="currentColor"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <path
                            fill-rule="evenodd"
                            clip-rule="evenodd"
                            d="M5.829 6.76a1.932 1.932 0 100-3.863 1.932 1.932 0 000 3.863zm0 2.898a4.829 4.829 0 100-9.658 4.829 4.829 0 000 9.658z"
                        />
                        <path
                            d="M22.473 1.867H30.2v7.726h-7.726V1.867zM22.473 13.07H30.2v7.727h-7.726V13.07zM22.473 24.274H30.2V32h-7.726v-7.726z"
                        />
                        <path
                            fill-rule="evenodd"
                            clip-rule="evenodd"
                            d="M12.014 5.795c0-.8.648-1.449 1.448-1.449h5.795a1.449 1.449 0 110 2.897h-5.795c-.8 0-1.448-.648-1.448-1.448zM6.544 11.047a1.449 1.449 0 00-1.6 1.28l1.44.16-1.44-.16v.011l-.003.023-.008.08c-.006.066-.015.162-.024.283-.018.242-.04.587-.055 1.013a28.23 28.23 0 00.087 3.36c.226 2.602.915 6.018 2.937 8.546 2.08 2.599 5.13 3.566 7.48 3.918a18.29 18.29 0 003.957.15c.111-.008.2-.017.263-.023l.076-.008.023-.003h.008l.003-.001s.002 0-.178-1.438l.18 1.438a1.449 1.449 0 00-.358-2.875M7.824 12.646l-.001.012-.006.055-.02.231a25.333 25.333 0 00.03 3.902c.212 2.43.835 5.14 2.314 6.987 1.42 1.776 3.62 2.56 5.646 2.863a15.408 15.408 0 003.303.127 7.78 7.78 0 00.193-.017l.043-.005h.006M6.544 11.046a1.449 1.449 0 011.28 1.6l-1.28-1.6z"
                        />
                        <path
                            fill-rule="evenodd"
                            clip-rule="evenodd"
                            d="M5.692 11.214a1.449 1.449 0 00-.58 1.965l1.272-.692-1.272.692v.002l.002.002.003.006.008.014.023.04a8.703 8.703 0 00.353.551 12.492 12.492 0 005.602 4.416c2.047.807 4.203 1.038 5.803 1.079a21.55 21.55 0 001.986-.04 16.55 16.55 0 00.742-.067l.047-.006.014-.002h.008l-.193-1.436.192 1.435a1.45 1.45 0 00-.383-2.871h-.002l-.027.003a13.35 13.35 0 01-.594.053c-.416.028-1.012.052-1.716.035-1.424-.037-3.205-.244-4.814-.878a9.594 9.594 0 01-4.286-3.373 5.756 5.756 0 01-.221-.345l-.005-.008a1.449 1.449 0 00-1.962-.575z"
                        />
                    </svg>
                    Batch Changes
                </a>
            </li>
            <li aria-current={$page.url.pathname === '/insights' ? 'page' : undefined}>
                <a href="/insights">
                    <svg aria-hidden="true" viewBox="0 0 25 25">
                        <path d={mdiChartBar} />
                    </svg>
                    Insights
                </a>
            </li>
            -->
        </ul>
    </nav>
    <div class="user">
        {#if authenticatedUser}
            <UserAvatar user={authenticatedUser} />
            <!--
                Needs data-sveltekit-reload to force hitting the server and
                proxying the request to the Sourcegraph instance.
            -->
            <a href="/-/sign-out" data-sveltekit-reload>Sign out</a>
        {:else}
            <a href="/sign-in" data-sveltekit-reload>Sign in</a>
        {/if}
    </div>
</header>

<style lang="scss">
    header {
        display: flex;
        align-items: center;
        border-bottom: 1px solid var(--border-color-2);
        height: var(--navbar-height);
        min-height: 40px;
        padding: 0 1rem;
        background-color: var(--color-bg-1);
    }

    img {
        &:hover {
            @keyframes spin {
                50% {
                    transform: rotate(180deg) scale(1.2);
                }
                100% {
                    transform: rotate(180deg) scale(1);
                }
            }

            @media (prefers-reduced-motion: no-preference) {
                animation: spin 0.5s ease-in-out 1;
            }
        }
    }

    nav {
        display: flex;
        align-self: stretch;
        flex: 1;
    }

    svg {
        width: 1rem;
        height: 1rem;
        margin-right: 0.5rem;
    }

    ul {
        position: relative;
        padding: 0;
        margin: 0;
        display: flex;
        justify-content: center;
        list-style: none;
        background-size: contain;
    }
</style>

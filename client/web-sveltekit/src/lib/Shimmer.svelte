
<div class="container">
    <div class="shimmer-animate" />
</div>

<style lang="scss">
    .container {
        height: var(--height, 1rem);
        display: flex;
        align-items: center;
        flex: 1;
    }

    :global(.theme-dark) {
        --shimmer-gradient: linear-gradient(90deg, var(--black) 100%, var(--gray-10) 100%);
    }
    :global(.theme-light) {
        --shimmer-gradient: linear-gradient(90deg, var(--gray-04) -1.57%, var(--gray-05) 100%);
    }

    .shimmer-animate{
        border-radius: 8px;
        height: 1rem;
        flex: 1;
        animation : shimmer 2s infinite;
        background: linear-gradient(to right, #eff1f3 4%, #e2e2e2 25%, #eff1f3 36%);
        background-size: 1000px 100%;
    }

    @keyframes shimmer {
      0% {
        background-position: -1000px 0;
      }
      100% {
        background-position: 1000px 0;
      }
    }
</style>

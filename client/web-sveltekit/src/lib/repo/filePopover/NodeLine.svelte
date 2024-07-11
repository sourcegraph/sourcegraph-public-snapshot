<div class="line">
    <div class="connector" />
    <div class="circle" />
    <div class="connector" />
</div>

<style lang="scss">
    $circle-size: 8px;

    .line {
        display: flex;
        flex-direction: column;
        align-items: center;
        height: 100%;

        .connector {
            border-left: 1px solid var(--border-color);
            flex-grow: 1;
        }

        .circle {
            height: $circle-size;
            width: $circle-size;
            border-radius: 50%;
            border: 2px solid var(--border-color);
            flex-grow: 0;
        }
    }
</style>

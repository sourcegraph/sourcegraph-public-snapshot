<!--
TODO: Make this component reuseable. It can only be used in the FilePopover at the moment.
-->
<script lang="ts">
</script>

<div class="line">
    <div class="connector" />
    <div class="circle" />
    <div class="connector" />
</div>

<style lang="scss">
    .line {
        display: flex;
        flex-flow: column nowrap;
        align-items: center;
        justify-content: center;

        .connector {
            border-left: 1px solid var(--border-color);
            height: 43px;
        }

        .circle {
            height: 6px;
            width: 6px;
            border-radius: 50%;
            border: 1.5px solid var(--border-color);
        }
    }
</style>

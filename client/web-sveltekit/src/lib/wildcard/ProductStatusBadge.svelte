<script lang="ts" context="module">
    import type { BADGE_VARIANTS } from './Badge.svelte'

    export const PRODUCT_STATUSES = ['beta', 'private beta', 'experimental', 'wip', 'new'] as const
    export type ProductStatusType = typeof PRODUCT_STATUSES[number]

    /**
     * Product statuses mapped to Badge style variants
     */
    const STATUS_VARIANT_MAPPING: Record<ProductStatusType, typeof BADGE_VARIANTS[number]> = {
        wip: 'warning',
        experimental: 'warning',
        beta: 'info',
        'private beta': 'info',
        new: 'info',
    }
</script>

<script lang="ts">
    import type { ComponentProps } from 'svelte'

    import Badge from './Badge.svelte'

    type $$Props = Omit<ComponentProps<Badge>, 'variant'> & {
        status: ProductStatusType
    }
    export let status: ProductStatusType
</script>

<span>
    <Badge variant={STATUS_VARIANT_MAPPING[status]} {...$$restProps}>
        {status}
    </Badge>
</span>

<style lang="scss">
    span {
        display: contents;
        text-transform: capitalize;
    }
</style>

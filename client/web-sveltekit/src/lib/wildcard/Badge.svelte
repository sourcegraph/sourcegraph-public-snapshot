<script lang="ts" context="module">
    import classNames from 'classnames'
    import styles from './Badge.module.scss'

    export const BADGE_VARIANTS = [
        'primary',
        'secondary',
        'success',
        'danger',
        'warning',
        'info',
        'merged',
        'outlineSecondary',
    ] as const

    export type BadgeVariantType = typeof BADGE_VARIANTS[number]

    export function badgeClassName(variant: BadgeVariantType, small?: boolean, pill?: boolean): string {
        return classNames(styles.badge, styles[variant], { [styles.small]: small, [styles.pill]: pill })
    }
</script>

<script lang="ts">
    /**
     * The variant style of the badge.
     */
    export let variant: BadgeVariantType
    /**
     * Allows modifying the size of the badge. Supports a smaller variant.
     */
    export let small: boolean | undefined = undefined
    /**
     * Render the badge as a rounded pill
     */
    export let pill: boolean | undefined = undefined

    $: cls = badgeClassName(variant, small, pill)
</script>

<!--TODO: support non-branded badges -->
<slot name="custom" class={cls}>
    <span class={cls}>
        <slot />
    </span>
</slot>

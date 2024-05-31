import classNames from 'classnames'

import { Button, Icon, LoadingSpinner, Text, type ButtonProps } from '@sourcegraph/wildcard'

import styles from './LoadingIconButton.module.scss'

/**
 * LoadingIconButton renders loading spinner or provided icon depending on the `isLoading` prop
 */
export const LoadingIconButton: React.FC<
    React.PropsWithChildren<ButtonProps & { isLoading: boolean; className?: string; iconSvgPath: string }>
> = ({ isLoading, className, iconSvgPath, children, ...buttonProps }) => (
    <Button {...buttonProps} className={classNames('d-flex align-items-center', className)}>
        {isLoading ? (
            <LoadingSpinner className={classNames('ml-0', styles.icon)} />
        ) : (
            <Icon aria-hidden={true} className={styles.icon} svgPath={iconSvgPath} />
        )}
        <Text as="span">{children}</Text>
    </Button>
)

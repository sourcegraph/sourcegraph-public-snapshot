import React, { HTMLAttributes } from 'react'

type AppRouterContainerProps = HTMLAttributes<HTMLDivElement>

export const AppRouterContainer: React.FunctionComponent<AppRouterContainerProps> = ({
    children,
    className,
    ...rest
}) =>
    children /* (
    <div className={classNames(styles.appRouterContainer, className)} {...rest}>
        {children}
    </div>
)*/

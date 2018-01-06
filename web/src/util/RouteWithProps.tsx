import * as React from 'react'
import { Route, RouteComponentProps, RouteProps } from 'react-router'

interface Props<O, M> extends RouteProps {
    component: React.ComponentType<ComponentProps<O, M>> // React.ComponentType<ComponentProps<O, M>>
    other?: O
}

type ComponentProps<O, M> = O & RouteComponentProps<M>

/**
 * A wrapper around react-router's Route that transfers other additional props
 * to a route's component.
 *
 * @template O is the type of the other additional props to transfer
 * @template M is the type of the route match data
 */
export const RouteWithProps = <O extends object, M = any>(
    props: Props<O, M>
): React.ReactElement<Route<Props<O, M>>> => (
    <Route
        {...props}
        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
        component={undefined}
        // tslint:disable-next-line:jsx-no-lambda
        render={props2 => {
            const finalProps: ComponentProps<O, M> = {
                ...props2,
                ...(props.other as object),
            } as ComponentProps<O, M> // cast needed until https://github.com/Microsoft/TypeScript/pull/13288 lands
            if (props.component) {
                const C = props.component
                return <C {...finalProps} />
            }
            if (props.render) {
                return props.render(finalProps)
            }
            return null
        }}
    />
)

import { RouteProps, Route as RouteV5 } from 'react-router-dom'
import { Route as RouteV6, Routes } from 'react-router-dom-v5-compat'

type Props = RouteProps & {
    pathV6: string
}

/**
 * Compatibility Route component to help us migrate from `react-router-dom` v5 to
 * v6. It renders a V5 route inside of a V6 route.
 *
 * It takes an intersection type of V5 RouteProps and a pathV6 prop.
 */
export function XCompatRoute({ pathV6, ...rest }: Props): JSX.Element {
    return (
        <Routes location={rest.location}>
            <RouteV6 path={pathV6} element={<RouteV5 {...rest} />} />
        </Routes>
    )
}

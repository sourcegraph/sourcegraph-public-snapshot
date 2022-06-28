import * as React from 'react'

import type { MdiReactIconProps } from 'mdi-react'

export const PerforceIcon: React.FunctionComponent<React.PropsWithChildren<MdiReactIconProps>> = ({
    color = 'currentColor',
    size = 24,
    className = '',
    ...props
}) => (
    <svg {...props} className={className} width={size} height={size} fill={color} viewBox="0 0 24 24">
        <path d="M3.742 8.754c.16-.418.352-.828.57-1.219l-.71-.644c2.773-3.325 6.39-4.32 9.59-3.743.656.09 1.308.247 1.956.485 4.582 1.703 6.903 6.754 5.18 11.285-.172.45-.387.883-.613 1.285.254.219.808.629.777.664-3.078 3.637-7.176 4.48-10.59 3.469-.328-.082-.652-.18-.98-.297-4.574-1.703-6.899-6.75-5.18-11.285zM19.372.98L17.75 2.512c-.54-.301-1.121-.582-1.727-.801C10.82-.227 5.336 1.965 2.316 6.03.738 8.363-.195 11.234.036 14.188c0 0 .007 5.558 5.136 8.832l1.305-1.786c.57.328 1.175.621 1.816.86 5.89 2.183 12.418-.606 14.555-6.23 0 0 1.562-3.43 1.047-7.177 0 0-.399-5.058-4.524-7.71zm0 0" />
    </svg>
)

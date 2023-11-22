import type { MdiReactIconComponentType } from 'mdi-react'

export const GerritIcon: MdiReactIconComponentType = props => (
    <svg
        className={'mdi-icon ' + (props.className || '')}
        width={props.size ?? 24}
        height={props.size ?? 24}
        fill={props.color ?? 'currentColor'}
        viewBox="0 0 52 52"
    >
        <path
            d="M4,0 h32 a4,4 0 0 1 4,4 v8 h8 a4,4 0 0 1 4,4 v32 a4,4 0 0 1 -4,4 h-32 a4,4 0 0 1 -4,-4 v-8 h-8 a4,4 0 0 1 -4,-4 v-32 a4,4 0 0 1 4,-4 z
         m14,22 h12 v4 h-12 v-4 z
         m16,0 h12 v4 h-12 v-4 z
         m-16,14 h4 v-4 h4 v4 h4 v4 h-4 v4 h-4 v-4 h-4 v-4 z
         m16,0 h4 v-4 h4 v4 h4 v4 h-4 v4 h-4 v-4 h-4 v-4 z"
            fillRule="evenodd"
        />
    </svg>
)

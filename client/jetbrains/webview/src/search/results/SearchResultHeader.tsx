import React from 'react'

interface Props {
    children: React.ReactNode
}

export const SearchResultHeader: React.FunctionComponent<Props> = ({ children }: Props) => <div>{children}</div>

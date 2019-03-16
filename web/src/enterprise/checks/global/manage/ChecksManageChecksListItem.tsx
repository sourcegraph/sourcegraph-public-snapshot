import React from 'react'
import { Check } from '../../data'

interface Props {
    node: Check
}

export const ChecksManageChecksListItem: React.FunctionComponent<Props> = ({ node }) => (
    <li className="list-group-item">{node.title}</li>
)

import * as React from 'react'
import { Check } from '../data'

interface Props {
    check: Check
}

/**
 * The manage page for a single check.
 */
export const CheckManagePage: React.FunctionComponent<Props> = ({ check }) => (
    <div className="check-manage-page">MANAGE</div>
)

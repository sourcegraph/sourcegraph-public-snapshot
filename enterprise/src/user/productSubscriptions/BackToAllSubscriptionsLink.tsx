import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'

export const BackToAllSubscriptionsLink: React.SFC<{ user: Pick<GQL.IUser, 'url'> }> = ({ user }) => (
    <Link to={`${user.url}/subscriptions`} className="btn btn-outline-link btn-sm mb-3">
        <ArrowLeftIcon className="icon-inline" /> All subscriptions
    </Link>
)

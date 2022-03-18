import React, { useEffect, useState } from 'react'

import GithubIcon from 'mdi-react/GithubIcon'
import { Button } from 'reactstrap'

import { Card, CardBody, PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'

const connectOrg = (installation_id: string) => () => {
    const queryString = window.location.search
    const urlParameters = new URLSearchParams(queryString)
    const state = urlParameters.get('state')
    window.location.assign(
        '/setup/github/app/cloud?setup_action=install&installation_id=' + installation_id + '&state=' + state
    )
}

export const ConnectGitHubAppPage: React.FunctionComponent<{}> = () => {
    const [data, setData] = useState([])

    useEffect(() => {
        fetch('/.auth/github/get-user-orgs', {
            method: 'GET',
        })
            .then(response => response.json())
            .then(response => {
                if (response.length === 0) {
                    console.log('redirect')
                } else {
                    setData(response)
                }
            })
            .catch(console.log)
    }, [])

    return (
        <Page>
            <PageTitle>Success!</PageTitle>
            <PageHeader
                path={[
                    {
                        icon: GithubIcon,
                        text: 'Complete your GitHub connection',
                    },
                ]}
                className="mb-3"
            />
            <Card>
                <CardBody>
                    {data.map(install => (
                        <>
                            <Button onClick={connectOrg(install.id)}>
                                <div className="d-flex align-items-start">
                                    <div className="align-self-center">
                                        <UserAvatar
                                            className="icon-inline mb-0 mr-1"
                                            user={{ avatarURL: install.account.avatar_url }}
                                        />
                                    </div>
                                    <div className="flex-1 align-self-center">
                                        <h3 className="m-0">{install.account.login}</h3>
                                    </div>
                                </div>
                            </Button>
                        </>
                    ))}
                </CardBody>
            </Card>
        </Page>
    )
}

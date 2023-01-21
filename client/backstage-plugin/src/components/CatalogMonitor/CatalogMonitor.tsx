import React, { useEffect, useState } from 'react';
import { Typography, Grid } from '@material-ui/core';
import {
  InfoCard,
  Header,
  Page,
  Content,
  ContentHeader,
  HeaderLabel,
  SupportButton,
} from '@backstage/core-components';
import { GraphQLClient, gql } from 'graphql-request';

interface Query<T> {
  gql(): string
  Marshal(data: string): T
}

export class SearchQuery implements Query<any> {
  Marshal(_: string): any {
    throw new Error('Method not implemented.');
  }
  gql(): string {
    throw new Error('Method not implemented.');
  }
}

class UserQuery implements Query<string> {
  Marshal(data: any): string {
    if ("currentUser" in data) {
      return data.currentUser.username;
    }
    throw new Error("username not found")
  }
  gql(): string {
    return gql`
    query {
      currentUser {
        username
      }
    }
    `
  }

}


class SourcegraphClient {
  private client: GraphQLClient

  constructor(baseUrl: string, token: string, sudo: boolean = false, sudoUsername: string = "") {
    const authz = sudo ? `token-sudo user="${sudoUsername}",token="${token}"` : `token ${token}`
    const apiUrl = `${baseUrl}/.api/graphql`
    this.client = new GraphQLClient(apiUrl,
      {
        headers: {
          'X-Requested-With': `Sourcegraph - Backstage plugin DEV`,
          Authorization: authz,
        }
      })
  }

  async ping(): Promise<string> {
    const q = new UserQuery()

    const data = await this.fetch(q)
    return data
  }

  async fetch<T>(q: Query<T>): Promise<T> {
    const data = await this.client.request(q.gql())

    return q.Marshal(data)
  }
}

const SUDO_TOKEN = "f385a9ebca8276e049eec289a521b50ab69a44aa"

export const CatalogMonitor = () => {
  const [stuff, setStuff] = useState('')
  const sg = new SourcegraphClient("https://sourcegraph.test:3443", SUDO_TOKEN, true, "sourcegraph")

  useEffect(() => {
    async function get() { setStuff(await sg.ping()) }
    get()
  })
  return (
    <Page themeId="tool">
      <Header title="Welcome to Sourcegraph!" subtitle="Code Intelligence Platform">
        <HeaderLabel label="Owner" value="DevX" />
        <HeaderLabel label="Lifecycle" value="Alpha" />
      </Header>
      <Content>
        <ContentHeader title="Catalog monitor">
          <SupportButton>Automatic load your software catalog</SupportButton>
        </ContentHeader>
        <Grid container spacing={3} direction="column">
          <Grid item>
            <InfoCard title="Information card">
              <Typography variant="body1">
                {stuff}
              </Typography>
            </InfoCard>
          </Grid>
        </Grid>
      </Content>
    </Page>
  )
}

import { Content, ContentHeader, Header, HeaderLabel, InfoCard, Page, SupportButton } from '@backstage/core-components'
import { Grid, Typography } from '@material-ui/core'
import React from 'react'
import { ExampleFetchComponent } from '../ExampleFetchComponent'

export const ExampleComponent = () => (
    <Page themeId="tool">
        <Header title="Welcome to Sourcegraph!" subtitle="Optional subtitle">
            <HeaderLabel label="Owner" value="Team X" />
            <HeaderLabel label="Lifecycle" value="Alpha" />
        </Header>
        <Content>
            <ContentHeader title="Plugin title">
                <SupportButton>A description of your plugin goes here.</SupportButton>
            </ContentHeader>
            <Grid container spacing={3} direction="column">
                <Grid item>
                    <InfoCard title="Information card">
                        <Typography variant="body1">All content should be wrapped in a card like this.</Typography>
                    </InfoCard>
                </Grid>
                <Grid item>
                    <ExampleFetchComponent />
                </Grid>
            </Grid>
        </Content>
    </Page>
)

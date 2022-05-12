import React, { FunctionComponent } from 'react'

import { Meta } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { Typography } from '@sourcegraph/wildcard'

import { LogoAscii } from '../LogoAscii'

import { Terminal, TerminalDetails, TerminalLine, TerminalProgress, TerminalTitle } from '.'

const cloningStatusLines = [
    {
        id: '1',
        title: 'sourcegraph/secret-project',
        details: '550kb of 2.5mb transferred ',
        progress: 22,
    },
    {
        id: '2',
        title: 'sourcegraph/test-repo',
        details: '11.25mb of 15mb transferred ',
        progress: 74,
    },
    {
        id: '3',
        title: 'sourcegraph/cloud',
        details: 'successfully cloned',
        progress: 100,
    },
]

export const InProgress: FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <>
        <div className="border overflow-hidden rounded">
            <header>
                <div className="py-3 px-4">
                    <Typography.H3 className="d-inline-block m-0">Activity log</Typography.H3>
                    <span className="float-right m-0">1/3 done</span>
                </div>
            </header>
            <Terminal>
                <TerminalLine>
                    <code>Cloning Repositories...</code>
                </TerminalLine>

                {cloningStatusLines?.map(({ id, title, details, progress }) => (
                    <React.Fragment key={id}>
                        <TerminalLine>
                            <TerminalTitle>{title}</TerminalTitle>
                        </TerminalLine>
                        <TerminalLine>
                            <TerminalDetails>{details}</TerminalDetails>
                        </TerminalLine>
                        <TerminalLine>
                            <TerminalProgress character="#" progress={progress} />
                        </TerminalLine>
                    </React.Fragment>
                ))}
            </Terminal>
        </div>
    </>
)

export const Finished: FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <>
        <div className="border overflow-hidden rounded">
            <header>
                <div className="py-3 px-4">
                    <Typography.H3 className="d-inline-block m-0">Activity log</Typography.H3>
                    <span className="float-right m-0">3/3 done</span>
                </div>
            </header>
            <Terminal>
                <TerminalLine>
                    <TerminalTitle>Done!</TerminalTitle>
                </TerminalLine>
                <TerminalLine>
                    <LogoAscii />
                </TerminalLine>
            </Terminal>
        </div>
    </>
)

const Story: Meta = {
    title: 'web/Terminal',

    decorators: [
        story => <BrandedStory styles={webStyles}>{() => <div style={{ width: 595 }}>{story()}</div>}</BrandedStory>,
    ],

    parameters: {
        component: Terminal,
    },
}

export default Story

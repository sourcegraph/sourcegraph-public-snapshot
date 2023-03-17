import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CodyChat } from './CodyChat'

/*
Some prompts:

You are an expert code reviewer. What specific changes would you make to improve the code in this file? List 3 at most. Do not make general suggestions; only make suggestions that are specific to the code in this file. An expert code reviewer will not suggest using longer variable names.

You are an expert penetration tester. What are the 3 most likely security vulnerabilities in this code file? Do not make general suggestions; only make suggetsions that are specific to the code in this file. Expert penetration testers check for OWASP top 10 violations, allow-by-default (instead of deny-by-default), logging or storing credentials in plain text.

You are a highly experienced staff engineer. Summarize what this file does.

You are a technical writer. Propose readability and clarity improvements to literal strings and error messages in this file. Retain useful detail in the error messages. Pay particular attention to typos, incorrect punctuation, bad grammar, and inconsistency.

Summarize the TODOs in this file. If a TODO is no longer valid, mention that it could be removed.

Write a doc comment for ...

*/

export const CodyPanel: React.FunctionComponent<
    {
        repoID: string
        repoName: string
        revision?: string
        filePath: string
        blobContent: string
    } & TelemetryProps
> = ({ repoName, filePath, blobContent, telemetryService }) => {
    useEffect(() => {
        telemetryService.log('CodyPanelOpened')
    }, [telemetryService])

    return (
        <div className="pt-3">
            <CodyChat
                promptPrefix={[
                    `Human: The content of the file '${filePath}' in repository '${repoName}' is`,
                    `<file>\n${blobContent}</file>`,
                    'Human: ',
                ].join('\n\n')}
            />
        </div>
    )
}

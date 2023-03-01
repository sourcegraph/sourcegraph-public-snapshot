import { useMemo } from 'react'

import { EditorState, Extension } from '@codemirror/state'
import { EditorView, lineNumbers } from '@codemirror/view'
import { mdiTrashCan, mdiUpload } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { CodeMirrorEditor, defaultEditorTheme } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { H3, Text, Button, Icon, Code, Alert } from '@sourcegraph/wildcard'

import { RepositoryFields } from '../../graphql-operations'

import { RepositoryOwnAreaPageProps } from './RepositoryOwnPage'
import { testCodeOwnersIngestedFile } from './testData'

import styles from './RepositoryOwnPageContents.module.scss'

export interface CodeownersIngestedFile {
    contents: string
    updatedAt: string
}

export const RepositoryOwnPageContents: React.FunctionComponent<
    Pick<RepositoryOwnAreaPageProps, 'repo' | 'authenticatedUser'>
> = ({ repo, authenticatedUser }) => {
    const isAdmin = authenticatedUser?.siteAdmin

    const codeownersIngestedFile: CodeownersIngestedFile | null = testCodeOwnersIngestedFile

    const isLightTheme = useIsLightTheme()

    const extensions: Extension[] = useMemo(
        () => [
            EditorView.darkTheme.of(isLightTheme === false),
            EditorState.readOnly.of(true),
            lineNumbers(),
            defaultEditorTheme,
        ],
        [isLightTheme]
    )

    return (
        <>
            <div className={styles.columns}>
                <div>
                    <H3>{isAdmin ? 'Upload a CODEOWNERS file' : 'Ask your site admin to upload a CODEOWNERS file'}</H3>
                    <Text>
                        Each owner must be either a Sourcegraph username, a Sourcegraph team name, or an email address.
                    </Text>

                    {isAdmin && (
                        <>
                            <Button variant="primary">
                                <Icon svgPath={mdiUpload} aria-hidden={true} className="mr-2" />
                                {codeownersIngestedFile ? 'Replace current file' : 'Upload file'}
                            </Button>
                            {codeownersIngestedFile && (
                                <Button variant="danger" className="ml-2">
                                    <Icon svgPath={mdiTrashCan} aria-hidden={true} className="mr-2" />
                                    Delete uploaded file
                                </Button>
                            )}
                        </>
                    )}
                </div>

                <div className={styles.or}>
                    <div className={styles.orLine} />
                    <div className="py-2">or</div>
                    <div className={styles.orLine} />
                </div>

                <div>
                    <H3>Commit a CODEOWNERS file</H3>
                    <Text>
                        Add a <Code>CODEOWNERS</Code> file to the root of this repository. Owners must be{' '}
                        {getCodeHostName(repo)} usernames or email addresses.
                    </Text>
                    {codeownersIngestedFile && (
                        <Alert className="mt-3" variant="info">
                            Since an uploaded file exists, any commited CODEOWNERS file in this repository will be
                            ignored.
                        </Alert>
                    )}
                </div>
            </div>

            {codeownersIngestedFile && (
                <div className="mt-5">
                    <H3>Uploaded CODEOWNERS file</H3>
                    <Text>
                        The following CODEOWNERS file was uploaded to Sourcegraph{' '}
                        <Timestamp date={codeownersIngestedFile.updatedAt} />.
                    </Text>
                    <CodeMirrorEditor value={codeownersIngestedFile.contents} extensions={extensions} />
                </div>
            )}
        </>
    )
}

const getCodeHostName = (repo: RepositoryFields): string => {
    const externalServiceKind = repo.externalURLs[0]?.serviceKind

    switch (externalServiceKind) {
        case 'GITHUB':
            return 'GitHub'
        case 'GITLAB':
            return 'GitLab'
        default:
            return 'code host'
    }
}

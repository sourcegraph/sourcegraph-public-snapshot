import {FC, useCallback, useEffect, useMemo, useState} from 'react'
import {PageTitle} from "../../../../components/PageTitle";
import {Button, Container, Link, PageHeader, Popover, Text} from "@sourcegraph/wildcard";
import {ButtonGroup} from "@sourcegraph/wildcard";
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import {ConfirmationModal} from "../../../insights/components/modals/ConfirmationModal";



export const CodyGeneralPage: ({
                                   authenticatedUser,
                                   repo,
                                   telemetryService
                               }: { authenticatedUser: any; queryPolicies?: any; repo: any; telemetryService: any }) => void = ({
                                                authenticatedUser,
                                                repo,
                                                telemetryService,
                                            }) => {



    const [enabled, setEnabled] = useState<boolean>(false)
    const [showPopup, setShowPopup] = useState<boolean>(false)
    const swap = () => {
        setEnabled(!enabled)
    }
    return (
        <>
            <PageTitle>
                General Cody settings
            </PageTitle>
            <PageHeader
                headingElement="h2"
                description={<>Rules that control keeping embeddings up-to-date.</>}
                actions={
                    authenticatedUser?.siteAdmin && (
                        <Button to="./new?type=head" variant="primary" as={Link}>
                            Create new {!repo && 'global'} policy
                        </Button>
                    )
                }
                className="mb-3"
            />
            <div>
                {/*<ButtonGroup>*/}
                {/*    <Button onClick={swap} disabled={enabled} variant="primary">Enable</Button>*/}
                {/*    <Button onClick={swap} disabled={!enabled} variant="primary">Disable</Button>*/}

                {/*</ButtonGroup>*/}
                <Button onClick={() => {
                    setShowPopup(true)

                }} variant="primary">
                    {!enabled ? 'Enable Cody' : 'Disable Cody'}
                </Button>
                <ConfirmationModal showModal={showPopup} onCancel={() => setShowPopup(false)} onConfirm={() => {
                    swap()
                    setShowPopup(false)
                }}>
                    Are you sure?
                </ConfirmationModal>
                {/*<Toggle*/}
                {/*    onToggle={value => {*/}

                {/*        swap()*/}
                {/*    }}*/}
                {/*    title={enabled ? 'Enabled' : 'Disabled'}*/}
                {/*    id="job-enabled"*/}
                {/*    value={enabled}*/}
                {/*    // aria-label={`Toggle ${job.name} job`}*/}
                {/*/>*/}
                <Text>Enabling Cody will enable all Cody features, including indexing repository embeddings. Embeddings indexing sends code to the 3rd party service (OpenAI) to perform embeddings on the codebase.</Text>
            </div>
        </>
    )

}

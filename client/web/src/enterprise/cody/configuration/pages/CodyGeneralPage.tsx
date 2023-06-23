import React, {useEffect, useState} from 'react'
import {PageTitle} from '../../../../components/PageTitle';
import {ErrorAlert, Link, PageHeader, Text} from '@sourcegraph/wildcard';
import {ConfirmationModal} from '../../../insights/components/modals/ConfirmationModal';
import {fetchSite, updateSiteConfiguration} from '../../../../site-admin/backend';
import * as jsonc from 'jsonc-parser';
import {isCodyEnabled} from '../../../../cody/isCodyEnabled';
import {LoaderButton} from '../../../../components/LoaderButton';
import {useNavigate} from "react-router-dom";
import {H3} from "@storybook/components";

interface CodyGeneralPageProps {
}

const isEnabledQuery = `
query cody {
  site {
    isCodyEnabled
  }
}
`

export const CodyGeneralPage: React.FunctionComponent = () => {

    const [enabled, setEnabled] = useState<boolean>(isCodyEnabled)
    const [showPopup, setShowPopup] = useState<boolean>(false)
    const [updateLoading, setUpdateLoading] = useState<boolean>(false)
    const [updateError, setUpdateError] = useState<Error>(undefined)

    const navigate = useNavigate()

    const toggleCody = async (): Promise<void> => {
        await fetchSite().toPromise().then(config => {
            const id = config.configuration.id
            const content = config.configuration.effectiveContents

            const defaultModificationOptions: jsonc.ModificationOptions = {
                formattingOptions: {
                    eol: '\n',
                    insertSpaces: true,
                    tabSize: 2,
                },
            }

            const modification = jsonc.modify(content, ['cody.enabled'], !enabled, defaultModificationOptions)
            const modifiedContent = jsonc.applyEdits(content, modification)
            setUpdateLoading(true)
            setUpdateError(undefined)
            updateSiteConfiguration(id, modifiedContent).toPromise()
                .then(() => {
                    navigate(0)
                })
                .catch(error => setUpdateError(error))
                .finally(() => setUpdateLoading(false))
        })
    }

    return (
        <>
            <div>
                <PageTitle title="Cody general settings"/>
                <PageHeader
                    headingElement="h2"
                    path={[{text: 'Cody general settings'}]}
                    className="mb-3"
                />
            </div>

            <div>
                <Text>
                    Cody is currently {enabled ? 'enabled' : 'disabled'}.
                </Text>
                <LoaderButton onClick={() => setShowPopup(true)} variant="primary" loading={updateLoading}
                              label={!enabled ? 'Enable Cody' : 'Disable Cody'} alwaysShowLabel={true}/>
                {updateError && <ErrorAlert prefix="Error updating Cody status" error={updateError}/>}
                <ConfirmationModal showModal={showPopup} onCancel={() => setShowPopup(false)} onConfirm={async () => {
                    setShowPopup(false)
                    await toggleCody()
                }}>
                    <Text>
                        Are you sure?
                    </Text>
                    <Text>
                        {!enabled ? 'Enabling Cody will transmit code to a third party LLM and embeddings provider.' : 'Disabling Cody will disable access to all users.'}
                    </Text>
                </ConfirmationModal>

                <div className="mt-3">
                    <Text>Enabling Cody will enable all Cody features, including indexing repository embeddings.
                        Embeddings
                        indexing sends code to the 3rd party service (OpenAI) to perform embeddings on the
                        codebase.
                    </Text>
                    <Text>
                        By default, all repositories will index embeddings. To configure embeddings policies, <Link
                        to="/site-admin/embeddings/configuration">click here</Link>.
                    </Text>
                </div>

            </div>
        </>
    )

}

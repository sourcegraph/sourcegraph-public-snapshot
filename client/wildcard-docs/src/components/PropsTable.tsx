import { useDynamicImport } from 'docusaurus-plugin-react-docgen-typescript/pkg/dist-src/hooks/useDynamicImport'
import * as React from 'react'

export const PropTable = ({ name }) => {
    const props = useDynamicImport(name)

    if (!props) {
        return null
    }

    return (
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Default Value</th>
                    <th>Required</th>
                    <th>Description</th>
                </tr>
            </thead>
            <tbody>
                {Object.keys(props).map(key => (
                    <tr key={key}>
                        <td>
                            <code>{key}</code>
                        </td>
                        <td>
                            <code>{props[key].type?.name}</code>
                        </td>
                        <td>{props[key].defaultValue && <code>{props[key].defaultValue.value}</code>}</td>
                        <td>{props[key].required ? 'Yes' : 'No'}</td>
                        <td>{props[key].description}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    )
}

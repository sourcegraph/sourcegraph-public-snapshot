import React from 'react'
import _ from 'lodash'
import { mdiDownload } from '@mdi/js'

import { useLazyQuery } from '@sourcegraph/http-client'
import { Button, Icon } from '@sourcegraph/wildcard'

interface IProps<IQuery, IVariables> {
    query: string
    variables: IVariables
    path: string
    fileName: string
}

export const ExportButton = <IQuery, IVariables>({ query, variables, path, fileName }: IProps<IQuery, IVariables>) => {
    const [fetchCSV] = useLazyQuery<IQuery, IVariables>(query, {
        variables,
        onCompleted: data => {
            let element = document.createElement('a')
            element.setAttribute('href', 'data:text/csv;charset=utf-8,' + encodeURIComponent(_.get(data, path)))
            element.setAttribute('download', `${fileName}.csv`)
            element.style.display = 'none'
            document.body.appendChild(element)
            element.click()
            document.body.removeChild(element)
        },
    })

    return (
        <Button variant="primary" onClick={() => fetchCSV()}>
            <Icon className="mr-1" color="var(--white)" svgPath={mdiDownload} size="sm" aria-label="Download icon" />{' '}
            Export
        </Button>
    )
}

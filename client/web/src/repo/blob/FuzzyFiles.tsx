import React from 'react'

interface FuzzyFilesProps {
    files: string[]
}
export const FuzzyFiles: React.FunctionComponent<FuzzyFilesProps> = props => {
    return (
        <>
          <h1>POOP</h1>
          <ul>
              {props.files.map(f => <li key={f}>{f}</li>)}
          </ul>
        </>
    )
}

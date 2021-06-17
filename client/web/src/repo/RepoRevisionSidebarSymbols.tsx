import * as H from 'history'
import React, { useState } from 'react'
import Parser from 'web-tree-sitter'

import { RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { Scalars } from '../graphql-operations'

interface Props extends Partial<RevisionSpec> {
    repoID: Scalars['ID']
    history: H.History
    location: H.Location
    /** The path of the file or directory currently shown in the content area */
    activePath: string
}

function getTextContents(props: Props): string {
    return `package database

import "github.com/keegancsmith/sqlf"

// LimitOffset specifies SQL LIMIT and OFFSET counts. A pointer to it is typically embedded in other options
// structs that need to perform SQL queries with LIMIT and OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query fragment ("LIMIT %d OFFSET %d") for use in SQL queries.
func (o *LimitOffset) SQL() *sqlf.Query {
	if o == nil {
		return &sqlf.Query{}
	}
	return sqlf.Sprintf("LIMIT %d OFFSET %d", o.Limit, o.Offset)
}
`
}

const TREE_SITTER = 'https://tree-sitter.github.io/tree-sitter.wasm'
const TREE_SITTER_GO = 'https://tree-sitter.github.io/tree-sitter-go.wasm'

const GO_QUERY = `
(
    (function_declaration
        name: (identifier) @definition.function) ;@function
)
`

async function haveFun(text: string): Promise<void> {
    await Parser.init()
    console.log('Downloading Go tree-sitter')
    const parser = new Parser()
    const goParser = await Parser.Language.load(TREE_SITTER_GO)
    parser.setLanguage(goParser)
    // const query = goParser.query(GO_QUERY)
    const tree = parser.parse(text)
    console.log(tree)
}
function getSymbols(text: string): string[] {
    haveFun(text)
    return ['LimitOffset', 'SQL']
}

export const RepoRevisionSidebarSymbols: React.FunctionComponent<Props> = props => {
    const text = getTextContents(props)
    const symbols = getSymbols(text)
    return (
        <ul>
            {symbols.map(symbol => (
                <li key={symbol}>{symbol}</li>
            ))}
        </ul>
    )
}

import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

export interface RefData {
	repo: string;
	version: string;
	file: string;
	line: number;
	column: number;
}

export function resolveGlobalReferences(ref: RefData): Promise<Array<GQL.IRefFields>> {
	const { version, file, line, column, repo } = ref;
	const query =
		`query Content($repo: String, $version: String, $file: String, $line: Int, $column: Int) {
				root {
					repository(uri: $repo) {
						commit(rev: $version) {
							commit {
								id
								file(path: $file) {
									name
									definition(line: $line, column: $column) {
										globalReferences {
											refLocation {
												startLineNumber
												startColumn
												endLineNumber
												endColumn
											}
											uri {
												scheme
												host
												query
												fragment
												path
											}
										}
									}
								}
							}
						}
					}
				}
			}`;
	return fetchGraphQLQuery(query, { repo, version, file, line, column }).then((data) => {
		if (!data.root.repository || !data.root.repository.commit.commit || !data.root.repository.commit.commit.file || !data.root.repository.commit.commit.file.definition) {
			return [];
		}
		return Promise.resolve(data.root.repository.commit.commit.file.definition.globalReferences || []);
	}).catch(err => err);
}

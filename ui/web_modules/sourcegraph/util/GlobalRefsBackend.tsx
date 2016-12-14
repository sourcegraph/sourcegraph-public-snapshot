import { fetchGraphQLQuery } from "sourcegraph/util/GraphQLFetchUtil";

export interface RefData {
	language: string;
	repo: string;
	version: string;
	file: string;
	line: number;
	column: number;
}

export function resolveGlobalReferences(ref: RefData): Promise<Array<GQL.IRefFields>> {
	const { version, file, line, column, repo, language } = ref;
	const query =
		`query Content($repo: String, $version: String, $file: String, $line: Int, $column: Int, $language: String) {
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
	return fetchGraphQLQuery(query, { repo, version, file, line, column, language }).then((data) => {
		if (!data.root.repository || !data.root.repository.commit.commit || !data.root.repository.commit.commit.file || !data.root.repository.commit.commit.file.definition) {
			return [];
		}
		return Promise.resolve(data.root.repository.commit.commit.file.definition.globalReferences || []);
	}).catch(err => err);
}

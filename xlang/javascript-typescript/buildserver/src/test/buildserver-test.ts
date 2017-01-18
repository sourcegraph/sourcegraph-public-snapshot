import * as utils from 'javascript-typescript-langserver/lib/test/test-utils';

import { BuildHandler } from "../buildhandler";

// forcing strict mode
import * as util from 'javascript-typescript-langserver/lib/util';
util.setStrict(true);

import { testWithLangHandler } from 'javascript-typescript-langserver/lib/test/language-server-test';

// Run language-server tests with build handler
testWithLangHandler(() => new BuildHandler());

// Run build-handler-specific tests
describe('LSP BuildHandler', function () {
	this.timeout(20000);

	describe('single package.json at root', function () {
		before(function (done: () => void) {
			utils.setUp(new BuildHandler(), {
				'package.json':
				'{\n\
					"name": "mypkg",\n\
					"version": "4.0.2",\n\
					"scripts": {\n\
						"preinstall": "echo preinstall should not run && exit 1",\n\
						"postinstall": "echo postinstall should not run && exit 1",\n\
						"install": "echo install should not run && exit 1"\n\
					},\n\
					"dependencies": {\n\
						"typescript": "2.1.1",\n\
						"diff": "3.0.1"\n\
					},\n\
					"devDependencies": {\n\
						"@types/diff": "0.0.31"\n\
					}\n\
				}\n',
				'a.ts': "import * as diff from 'diff';\n\
import { diffChars, IDiffResult } from 'diff';\n\
\n\
diffChars('foo', 'bar');\n\
",
				'b.ts': "import * as ts from 'typescript';\n\
\n\
var s ts.SyntaxKind;\n\
var t = s;\n\
",
			}, done);
		});
		after(function (done: () => void) {
			utils.tearDown(done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 12
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 8,
							character: 0
						},
						end: {
							line: 86,
							character: 1
						}
					}
				}, done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 23
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 5,
							character: 0
						},
						end: {
							line: 87,
							character: 0
						}
					}
				}, done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 1,
					character: 10
				}
			}, [{
				uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
				range: {
					start: {
						line: 55,
						character: 4
					},
					end: {
						line: 55,
						character: 67
					}
				}
			}, {
				uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
				range: {
					start: {
						line: 55,
						character: 4
					},
					end: {
						line: 55,
						character: 67
					}
				}
			}], done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 1,
					character: 21
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 9,
							character: 4
						},
						end: {
							line: 14,
							character: 5
						}
					}
				}, done);
		})
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 1,
					character: 40
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 5,
							character: 0
						},
						end: {
							line: 87,
							character: 0
						}
					}
				}, done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 3,
					character: 0
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 55,
							character: 4
						},
						end: {
							line: 55,
							character: 67
						}
					}
				}, done);
		});
		it('cross-repo xdefinition', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 12
				}
			}, {
					symbol: {
						containerKind: "",
						containerName: "",
						kind: "module",
						name: "JsDiff",
						package: {
							name: "@types/diff",
							version: "0.0.31",
						},
					},
				}, done);
		});
		it('cross-repo xdefinition', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 23
				}
			}, {
					symbol: {
						containerKind: "",
						containerName: "",
						kind: "module",
						name: "\"/node_modules/@types/diff/index\"",
						package: {
							name: "@types/diff",
							version: "0.0.31",
						},
					},
				}, done);
		});
		it('cross-repo xdefinition', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 1,
					character: 10
				}
			}, [{
				symbol: {
					containerKind: "",
					containerName: "diff",
					kind: "function",
					name: "diffChars",
					package: {
						name: "@types/diff",
						version: "0.0.31",
					},
				},
			}, {
				symbol: {
					containerKind: "",
					containerName: "diff",
					kind: "function",
					name: "diffChars",
					package: {
						name: "@types/diff",
						version: "0.0.31",
					},
				},
			}], done);
		});
		it('cross-repo xdefinition', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 1,
					character: 21
				}
			}, {
					symbol: {
						containerKind: "",
						containerName: "diff",
						kind: "interface",
						name: "IDiffResult",
						package: {
							name: "@types/diff",
							version: "0.0.31",
						},
					},
				}, done);
		});
		it('cross-repo xdefinition', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 1,
					character: 40
				}
			}, {
					symbol: {
						containerKind: "",
						containerName: "",
						kind: "module",
						name: "\"/node_modules/@types/diff/index\"",
						package: {
							name: "@types/diff",
							version: "0.0.31",
						},
					},
				}, done);
		});
		it('cross-repo xdefinition', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 3,
					character: 0
				}
			}, {
					symbol: {
						containerKind: "",
						containerName: "JsDiff",
						kind: "function",
						name: "diffChars",
						package: {
							name: "@types/diff",
							version: "0.0.31",
						},
					},
				}, done);
		})
		it('cross-repo xdefinition to non-DefinitelyTyped package', async function (done: (err?: Error) => void) {
			utils.xdefinition({
				textDocument: {
					uri: 'file:///b.ts'
				},
				position: {
					line: 2,
					character: 9
				}
			}, {
					symbol: {
						containerKind: "",
						containerName: "ts",
						kind: "enum",
						name: "SyntaxKind",
						package: {
							name: "typescript",
							version: "2.1.1",
						},
					},
				}, done);
		});
		it('workspace/xreferences', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 12
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 8,
							character: 0
						},
						end: {
							line: 86,
							character: 1
						}
					}
				}, () => {
					utils.workspaceReferences({
						query: {
							containerKind: "",
							containerName: "diff",
							kind: "function",
							name: "diffChars",
							package: {
								name: "@types/diff",
								version: "0.0.31",
							},
						}
					}, [{
						reference: {
							range: {
								end: {
									character: 18,
									line: 1,
								},
								start: {
									character: 8,
									line: 1,
								},
							},
							uri: "file:///a.ts",
						},
						symbol: {
							containerKind: "",
							containerName: "diff",
							kind: "function",
							name: "diffChars",
							package: {
								name: "@types/diff",
								version: "0.0.31",
							},
						},
					}, {
						reference: {
							range: {
								end: {
									character: 18,
									line: 1,
								},
								start: {
									character: 8,
									line: 1,
								},
							},
							uri: "file:///a.ts",
						},
						symbol: {
							containerKind: "",
							containerName: "diff",
							kind: "function",
							name: "diffChars",
							package: {
								name: "@types/diff",
								version: "0.0.31",
							},
						},
					}], done);
				});
		});
	});

	describe('multiple package.json', function () {
		before(function (done: () => void) {
			utils.setUp(new BuildHandler(), {
				'package.json':
				'{\n\
					"name": "rootpkg",\n\
					"version": "4.0.2",\n\
					"dependencies": {\n\
						"diff": "3.0.1"\n\
					},\n\
					"devDependencies": {\n\
						"@types/diff": "0.0.31"\n\
					}\n\
				}\n',
				'a.ts': "import * as diff from 'diff';",
				'foo': {
					'b.ts': "import * as resolve from 'resolve';",
					'package.json':
					'{\n\
					"name": "foopkg",\n\
					"version": "1.0.3",\n\
					"dependencies": {\n\
						"resolve": "1.1.7"\n\
					},\n\
					"devDependencies": {\n\
						"@types/resolve": "0.0.4"\n\
					}\n\
				}\n',
				},
			}, done);
		});
		after(function (done: () => void) {
			utils.tearDown(done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 12
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#diff/index.d.ts',
					range: {
						start: {
							line: 8,
							character: 0
						},
						end: {
							line: 86,
							character: 1
						}
					}
				}, done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///foo/b.ts'
				},
				position: {
					line: 0,
					character: 26
				}
			}, {
					uri: 'git://github.com/DefinitelyTyped/DefinitelyTyped#resolve/index.d.ts',
					range: {
						start: {
							line: 13,
							character: 0
						},
						end: {
							line: 100,
							character: 0
						}
					}
				}, done);
		});
	});

	describe('vendored dependencies', function () {
		before(function (done: () => void) {
			utils.setUp(new BuildHandler(), {
				'package.json':
				'{\n\
					"name": "rootpkg",\n\
					"version": "4.0.2",\n\
					"dependencies": {\n\
						"diff": "1.0.1"\n\
					}\n\
				}\n',
				'a.ts': "import { x } from 'diff';",
				'node_modules': {
					'diff': {
						'index.d.ts': "export const x = 1;",
					},
				}
			}, done);
		});
		after(function (done: () => void) {
			utils.tearDown(done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 9
				}
			}, {
					uri: 'file:///node_modules/diff/index.d.ts',
					range: {
						start: {
							line: 0,
							character: 13
						},
						end: {
							line: 0,
							character: 18
						}
					}
				}, done);
		});
	});

	describe('dependency installation should not run scripts (javascript-dep-npm\'s scripts will fail)', function () {
		before(function (done: () => void) {
			utils.setUp(new BuildHandler(), {
				'package.json':
				'{\n\
					"name": "rootpkg",\n\
					"version": "4.0.2",\n\
					"dependencies": {\n\
						"javascript-dep-npm": "https://github.com/sgtest/javascript-dep-npm"\n\
					}\n\
				}\n',
				'a.ts': "import * as xyz from 'javascript-dep-npm';",
			}, done);
		});
		after(function (done: () => void) {
			utils.tearDown(done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 12
				}
			}, {
					uri: 'git://github.com/sgtest/javascript-dep-npm#index.d.ts',
					range: {
						start: {
							line: 0,
							character: 0
						},
						end: {
							line: 1,
							character: 0
						}
					}
				}, done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			utils.definition({
				textDocument: {
					uri: 'file:///a.ts'
				},
				position: {
					line: 0,
					character: 24
				}
			}, {
					uri: 'git://github.com/sgtest/javascript-dep-npm#index.d.ts',
					range: {
						start: {
							line: 0,
							character: 0
						},
						end: {
							line: 1,
							character: 0
						}
					}
				}, done);
		});
	});
});

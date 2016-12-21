import * as utils from 'javascript-typescript-langserver/src/test/test-utils';

import { BuildHandler } from "../buildhandler";

// forcing strict mode
import * as util from 'javascript-typescript-langserver/src/util';
util.setStrict(true);

import { testWithLangHandler } from 'javascript-typescript-langserver/src/test/language-server-test';

// Run language-server tests with build handler
testWithLangHandler(() => new BuildHandler());

// Run build-handler-specific tests
describe('LSP', function () {
	this.timeout(10000);
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
			}, done);
		});
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			try {
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
					}], err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
			} catch (e) {
				done(e);
				return;
			}
			done();
		});
		afterEach(function (done: () => void) {
			utils.tearDown(done);
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
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			try {
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
			} catch (e) {
				done(e);
				return;
			}
			done();
		});
		afterEach(function (done: () => void) {
			utils.tearDown(done);
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
		afterEach(function (done: () => void) {
			utils.tearDown(done);
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
		it('cross-repo definition', async function (done: (err?: Error) => void) {
			try {
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
				await new Promise<void>((resolve, reject) => {
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
						}, err => err ? reject(err) : resolve());
				});
			} catch (e) {
				done(e);
				return;
			}
			done();
		});
		afterEach(function (done: () => void) {
			utils.tearDown(done);
		});
	});
});

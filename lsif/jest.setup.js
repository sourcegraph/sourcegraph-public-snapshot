// LSIF tests create and migrate Postgres databases, which can take more
// time than the default test timeout. Increase it here for all tests in
// that project.

jest.setTimeout(15000)

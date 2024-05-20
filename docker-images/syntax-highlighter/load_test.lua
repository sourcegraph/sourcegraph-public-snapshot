-- Example Lua script to use with 'wrk' (https://github.com/wg/wrk)
-- for load testing the highlighter
wrk.method = "POST"
wrk.body   = '{"engine":"tree-sitter", "code": "int x=3;\\n", "filepath": "a.c", "filetype": "C"}'
wrk.headers["Content-Type"] = "application/json"

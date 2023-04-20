# Cody CLI (experimental)

Cody CLI is an experimental CLI of Cody.

## Install

In the root of the repository, run this:

```
pnpm --filter @sourcegraph/cody-cli run build
cd client/cody-cli
npm install -g .
```

Then you can do

```
cody
```

## Local Development

In the root of the repository, run this:

```
pnpm --filter @sourcegraph/cody-cli run start
```

## LSP

### Start as LSP

```
cody --lsp --stdio
```

### Neovim config

Rough, **hacky** Lua snippet to get `cody --lsp --stdio` running as `codylsp` in
Neovim:

```lua
vim.api.nvim_create_autocmd({ "FileType" }, {
  pattern = "*",
  callback = function()
    vim.lsp.start {
      name = "codylsp",
      cmd = { "cody", "--lsp", "--stdio" },
      root_dir = vim.fs.dirname(vim.fs.find({ "go.mod", ".git" }, { upward = true })[1]),
      capabilities = capabilities,
      on_attach = on_attach,
      trace = "off",
      settings = {
        codylsp = {
          sourcegraph = {
            url = "https://sourcegraph.sourcegraph.com",
            accessToken = "<your token>",
            repos = { "github.com/sourcegraph/sourcegraph" }, -- any repos you want context for
          },
        },
      },
    }
  end,
})

-- Define the "CodyR" ("cody replace") command that replaces the selection with whatever cody returns
vim.api.nvim_create_user_command("CodyR", function(command)
  local p = "file://" .. vim.fn.expand "%:p"

  for _, client in pairs(vim.lsp.get_active_clients { name = "codylsp" }) do
    client.request("workspace/executeCommand", {
      command = "cody.replace",
      arguments = { p, command.line1 - 1, command.line2 - 1, command.args },
    }, function() end, 0)
  end
end, { range = 2, nargs = 1 })

-- Define the "Cody" command that takes a question/prompt and shows the response
-- in popup window. If line/range is selected that's passed to Cody.
vim.api.nvim_create_user_command("Cody", function(command)
  local p = "file://" .. vim.fn.expand "%:p"

  for _, client in pairs(vim.lsp.get_active_clients { name = "codylsp" }) do
    client.request("workspace/executeCommand", {
      command = "cody.explain",
      arguments = { p, command.line1 - 1, command.line2 - 1, command.args },
    }, function(_, result, _, _)
      local lines = vim.split(result.response, "\n")
      vim.lsp.util.open_floating_preview(lines, "markdown", {
        height = 5,
        width = 80,
        focus_id = "codyResponse",
      })
      -- Call it again so that it focuses the window immediately
      vim.lsp.util.open_floating_preview(lines, "markdown", {
        height = 5,
        width = 80,
        focus_id = "codyResponse",
      })
    end, 0)
  end
end, { range = 2, nargs = 1 })

```

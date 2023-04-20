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
            accessToken = "sgp_5ac15c9b677628fcc7199d5fa42dca57528ea4a4",
            repos = { "github.com/sourcegraph/sourcegraph" }, -- any repos you want context for
          },
        },
      },
    }
  end,
})

vim.api.nvim_create_user_command("CodyR", function(command)
  local p = "file://" .. vim.fn.expand "%:p"

  for _, client in pairs(vim.lsp.get_active_clients { name = "codylsp" }) do
    client.request("workspace/executeCommand", {
      command = "cody",
      arguments = { p, command.line1 - 1, command.line2 - 1, command.args, true, true },
    }, function() end, 0)
  end
end, { range = 2, nargs = 1 })

vim.api.nvim_create_user_command("CodyC", function(command)
  local p = "file://" .. vim.fn.expand "%:p"

  for _, client in pairs(vim.lsp.get_active_clients { name = "codylsp" }) do
    client.request("workspace/executeCommand", {
      command = "cody",
      arguments = { p, command.line1 - 1, command.line2 - 1, command.args, false, true },
    }, function() end, 0)
  end
end, { range = 2, nargs = 1 })

vim.api.nvim_create_user_command("CodyE", function(command)
  local p = "file://" .. vim.fn.expand "%:p"

  for _, client in pairs(vim.lsp.get_active_clients { name = "codylsp" }) do
    client.request("workspace/executeCommand", {
      command = "cody.explain",
      arguments = { p, command.line1 - 1, command.line2 - 1, command.args },
    }, function(_, result, _, _)
      local lines = vim.split(result.message[1], "\n")
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

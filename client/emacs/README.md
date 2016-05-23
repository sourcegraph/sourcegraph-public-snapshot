# Emacs mode for Sourcegraph

Displays
[live usage examples and documentation for Go code](https://sourcegraph.com/github.com/golang/go@eb69476c66339ca494f98e65a78d315da99a9c79/-/def/GoPackage/net/http/-/Client/Get/-/info)
as you type, in [Emacs](https://www.gnu.org/software/emacs/).

## Install

sourcegraph-mode requires Emacs version 24.3+ and
[godefinfo](https://github.com/sqs/godefinfo).

``` shell
go get -u github.com/sqs/godefinfo
# ensure godefinfo is in your $PATH
godefinfo -v # should print "godefinfo version ___"
```

Place `sourcegraph.el` and `sourcegraph-autoloads.el` in any
directory, add that directory to your load path, and require
`'sourcegraph-autoloads`:

``` emacs-lisp
(add-to-list 'load-path "~/.emacs.d/personal")
(require 'sourcegraph)
(add-hook 'go-mode-hook 'sourcegraph-mode)
```

Evaluate these statements with `C-x C-e` or restart Emacs.

Now, open a `.go` file. When you put the cursor over any identifier,
your web browser will display context about it.

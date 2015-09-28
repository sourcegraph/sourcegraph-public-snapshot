## Starting background processes

For the lazy, the following function will launch all of the
applications needed to set up your dev environment in Emacs.


(defun src-start-work ()
  "Open up processes needed for work."
  (interactive)
  (let ((serve-dev (get-buffer-process (shell "*shell*<serve-dev>")))
        (webpack-dev-server (get-buffer-process (shell "*shell*<webpack-dev-server>")))
        (vcsstore (get-buffer-process (shell "*shell*<vcsstore>"))))
    (process-send-string serve-dev "cd $GOPATH/src/src.sourcegraph.com/sourcegraph\ngit pull\nmake serve-dev\n")
    (process-send-string webpack-dev-server "cd $GOPATH/src/src.sourcegraph.com/sourcegraph/app\nnpm start\n")
    (process-send-string vcsstore "cd $GOPATH/src/github.com/sourcegraph/vcsstore/\ndocker run -e GOMAXPROCS=8 -p 9090:80 -v /tmp/vcsstore vcsstore\n")))

# web-mode tab indentation

Use the following in your .emacs.d:

```
(setq web-mode-enable-tab-indentation t)
(setq-default indent-tabs-mode t)
(setq-default tab-width 4)
```

# Go template indentation

If you use [web-mode](http://web-mode.org/), add the following to your
.dir-locals.el (at the root of this repository) to make web-mode treat
.html files as Go template files.

```
((nil . (
         (setq web-mode-engines-alist '(("go" . "\\.html\\'")))
)))
```

# Fast projectile file finder

If you use [projectile](http://batsov.com/projectile/), add the
following to your .dir-locals.el to omit uninteresting files from the
list that `projectile-find-file` searches. This speeds it up
considerably.

((nil . (
         (projectile-git-command . "git ls-files -co --exclude-standard | grep -v Godeps/ | grep -v bower_components/ | grep -v node_modules/ | grep -v '\.test$' | grep -v vendor/ | tr '\\n' '\\0'")
)))

# web-mode jsx

Consider using the following snippets for better JSX support in
web-mode:
https://github.com/al3x/emacs/blob/e70f10c252f718ae77b39a447fded75d849c2401/mode-inits/init-webmode.el.

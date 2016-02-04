// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package semver provides Semantic Versioning for Go packages.
//
// For more information about semver please see:
//
// http://semver.org/
//
// This package allows for implementing semantic versioning of Go packages on
// a single GitHub user or organization under a custom domain, for example:
//
//  go get example.com/pkg.v1
//
// Would contact the Go HTTP server (using this package) running at example.com
// which would redirect the request to clone the Git repository located at:
//
//  github.com/someuser/pkg @ branch/tag [v1][v1.N][v1.N.N]
//
// Usage is pretty simple, first create a Handler with your configuration:
//
//  // Create a semver HTTP handler:
//  pkgHandler := &semver.Handler{
//      Host: "example.com",
//      Matcher: semver.GitHub("someuser"),
//  }
//
// Then register a root ("/") HTTP handler:
//
//  http.HandleFunc("/", handler)
//
// Inside of the root HTTP handler give the semver HTTP handler a chance to
// handle the request if it needs to:
//
//  func handler(w http.ResponseWriter, r *http.Request) {
//      // Give our semver handler the ability to handle the request.
//      status, err := pkgHandler.Handle(w, r)
//      if err != nil {
//          log.Println(err) // e.g. IO error
//      }
//      if status == semver.Handled {
//          // The request was handled by our semver pkgHandler, we don't need
//          // to do anything else.
//          return
//      }
//      if status == semver.PkgPage {
//          // Package page, redirect them to godoc.org documentation.
//          tmp := *r.URL
//          tmp.Scheme = "https"
//          tmp.Host = "godoc.org"
//          tmp.Path = path.Join(pkgHandler.Host, tmp.Path)
//          http.Redirect(w, r, tmp.String(), http.StatusSeeOther)
//          return
//      }
//
//      // It's not a package request -- do something else (e.g. render the
//      // home page).
//  }
//
// The package exposes a matcher only for GitHub. But others can be implemented
// outside the package as well for e.g. Google Code or privately hosted Git
// repositories.
package semver

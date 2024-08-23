// Copyright 2016 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package plural provides simple support for localising plurals in a flexible range of different styles.
//
// There are considerable differences around the world in the way plurals are handled. This is
// a simple but competent API for catering with these differences when presenting to people formatted text with numbers.
//
// This package is able to format countable things and continuous values. It can handle integers
// and floating point numbers equally and this allows you to decide to what extent each is appropriate.
//
// For example, "2 cars" might weigh "1.6 tonnes"; both categories are covered.
//
// This API is deliberately simple; it doesn't address the full gamut of internationalisation. If that's
// what you need, you should consider products such as https://github.com/nicksnyder/go-i18n instead.
//
// Please see the examples and associated api documentation.
//
package plural

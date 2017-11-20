// Package jsoncommentstrip contains functions which strips comments from JSON input.
//
// Supported comment types:
//
// - Single line:
//
//   {
//     // this is a single line comment
//   }
//
// - Multiple lines:
//
//   {
//     /* multiple
//      * lines comments
//      * works
//      */
//   }
//
// Supported line breaks:
//
//  \n    - unix style
//  \r\n  - windows style
//
// Supported escaping in JSON:
//     {
//         "test": "\"valid string // /*"
//     }
//
package jsoncommentstrip

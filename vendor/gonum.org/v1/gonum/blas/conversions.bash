#!/usr/bin/env bash

# Copyright ©2017 The Gonum Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Generate code for blas32.
echo Generating blas32/conv.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > blas32/conv.go
cat blas64/conv.go \
| gofmt -r 'float64 -> float32' \
\
| sed -e 's/blas64/blas32/' \
\
>> blas32/conv.go

echo Generating blas32/conv_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > blas32/conv_test.go
cat blas64/conv_test.go \
| gofmt -r 'float64 -> float32' \
\
| sed -e 's/blas64/blas32/' \
      -e 's_"math"_math "gonum.org/v1/gonum/internal/math32"_' \
\
>> blas32/conv_test.go

echo Generating blas32/conv_symmetric.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > blas32/conv_symmetric.go
cat blas64/conv_symmetric.go \
| gofmt -r 'float64 -> float32' \
\
| sed -e 's/blas64/blas32/' \
\
>> blas32/conv_symmetric.go

echo Generating blas32/conv_symmetric_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > blas32/conv_symmetric_test.go
cat blas64/conv_symmetric_test.go \
| gofmt -r 'float64 -> float32' \
\
| sed -e 's/blas64/blas32/' \
      -e 's_"math"_math "gonum.org/v1/gonum/internal/math32"_' \
\
>> blas32/conv_symmetric_test.go


# Generate code for cblas128.
echo Generating cblas128/conv.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas128/conv.go
cat blas64/conv.go \
| gofmt -r 'float64 -> complex128' \
\
| sed -e 's/blas64/cblas128/' \
\
>> cblas128/conv.go

echo Generating cblas128/conv_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas128/conv_test.go
cat blas64/conv_test.go \
| gofmt -r 'float64 -> complex128' \
\
| sed -e 's/blas64/cblas128/' \
      -e 's_"math"_math "math/cmplx"_' \
\
>> cblas128/conv_test.go

echo Generating cblas128/conv_symmetric.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas128/conv_symmetric.go
cat blas64/conv_symmetric.go \
| gofmt -r 'float64 -> complex128' \
\
| sed -e 's/blas64/cblas128/' \
\
>> cblas128/conv_symmetric.go

echo Generating cblas128/conv_symmetric_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas128/conv_symmetric_test.go
cat blas64/conv_symmetric_test.go \
| gofmt -r 'float64 -> complex128' \
\
| sed -e 's/blas64/cblas128/' \
      -e 's_"math"_math "math/cmplx"_' \
\
>> cblas128/conv_symmetric_test.go

echo Generating cblas128/conv_hermitian.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas128/conv_hermitian.go
cat blas64/conv_symmetric.go \
| gofmt -r 'float64 -> complex128' \
\
| sed -e 's/blas64/cblas128/' \
      -e 's/Symmetric/Hermitian/g' \
      -e 's/a symmetric/an Hermitian/g' \
      -e 's/symmetric/hermitian/g' \
      -e 's/Sym/Herm/g' \
\
>> cblas128/conv_hermitian.go

echo Generating cblas128/conv_hermitian_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas128/conv_hermitian_test.go
cat blas64/conv_symmetric_test.go \
| gofmt -r 'float64 -> complex128' \
\
| sed -e 's/blas64/cblas128/' \
      -e 's/Symmetric/Hermitian/g' \
      -e 's/a symmetric/an Hermitian/g' \
      -e 's/symmetric/hermitian/g' \
      -e 's/Sym/Herm/g' \
      -e 's_"math"_math "math/cmplx"_' \
\
>> cblas128/conv_hermitian_test.go


# Generate code for cblas64.
echo Generating cblas64/conv.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas64/conv.go
cat blas64/conv.go \
| gofmt -r 'float64 -> complex64' \
\
| sed -e 's/blas64/cblas64/' \
\
>> cblas64/conv.go

echo Generating cblas64/conv_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas64/conv_test.go
cat blas64/conv_test.go \
| gofmt -r 'float64 -> complex64' \
\
| sed -e 's/blas64/cblas64/' \
      -e 's_"math"_math "gonum.org/v1/gonum/internal/cmplx64"_' \
\
>> cblas64/conv_test.go

echo Generating cblas64/conv_hermitian.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas64/conv_hermitian.go
cat blas64/conv_symmetric.go \
| gofmt -r 'float64 -> complex64' \
\
| sed -e 's/blas64/cblas64/' \
      -e 's/Symmetric/Hermitian/g' \
      -e 's/a symmetric/an Hermitian/g' \
      -e 's/symmetric/hermitian/g' \
      -e 's/Sym/Herm/g' \
\
>> cblas64/conv_hermitian.go

echo Generating cblas64/conv_hermitian_test.go
echo -e '// Code generated by "go generate gonum.org/v1/gonum/blas”; DO NOT EDIT.\n' > cblas64/conv_hermitian_test.go
cat blas64/conv_symmetric_test.go \
| gofmt -r 'float64 -> complex64' \
\
| sed -e 's/blas64/cblas64/' \
      -e 's/Symmetric/Hermitian/g' \
      -e 's/a symmetric/an Hermitian/g' \
      -e 's/symmetric/hermitian/g' \
      -e 's/Sym/Herm/g' \
      -e 's_"math"_math "gonum.org/v1/gonum/internal/cmplx64"_' \
\
>> cblas64/conv_hermitian_test.go

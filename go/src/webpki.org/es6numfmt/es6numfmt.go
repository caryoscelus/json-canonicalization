//
//  Copyright 2006-2019 WebPKI.org (http://webpki.org).
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
//
 
// This package converts numbers in IEEE-754 double precision into the
// format specified for JSON in EcmaScript Version 6 and forward.
// The core application for this is canonicalization:
// https://tools.ietf.org/html/draft-rundgren-json-canonicalization-scheme-02

package es6numfmt

import (
    "errors"
    "math"
    "strconv"
    "strings"
)

const (
    allButSign uint64 = 0x7fffffffffffffff
    invalidNumber uint64 = 0x7ff0000000000000
)

func Convert(ieeeF64 float64) (res string, err error) {
    ieeeU64 := math.Float64bits(ieeeF64)

    // Special case: NaN and Infinity are invalid in JSON
    if (ieeeU64 & invalidNumber) == invalidNumber {
        return "null", errors.New("Invalid JSON number: " + strconv.FormatUint(ieeeU64, 16))
    }

    // Special case: eliminate "-0" which does not exist in JSON.  Takes "0" as well
    if (ieeeU64 & allButSign) == 0 {
        return "0", nil
    }

    // Deal with the sign separately
    var sign string = ""
    if ieeeF64 < 0 {
        ieeeF64 =-ieeeF64
        sign = "-"
    }

    // ES6 has a unique "g" format
    var format byte = 'g'
    if ieeeF64 < 1e+21 && ieeeF64 >= 1e-6 {
        format = 'f'
    }

    // The following should (in "theory") do the trick:
    es6Formatted := strconv.FormatFloat(ieeeF64, format, -1, 64)

    // Unfortunately Go version 1.14.4 is a bit buggy with respect to
    // rounding for -1 precision which is dealt with below.
    // https://github.com/golang/go/issues/29491
    exponent := strings.IndexByte(es6Formatted, 'e')
    if exponent > 0 {
        gform := strconv.FormatFloat(ieeeF64, 'g', 17, 64)
        if len(gform) == len(es6Formatted) {
           es6Formatted = gform
        }
        // Go outputs "1e+09" which must be rewritten as "1e+9"
        if es6Formatted[exponent + 2] == '0' {
            es6Formatted = es6Formatted[0:exponent + 2] + es6Formatted[exponent + 3:]
        }
    } else if strings.IndexByte(es6Formatted, '.') < 0 && len(es6Formatted) >= 12 {
        i := len(es6Formatted)
        for es6Formatted[i - 1] == '0' {
            i--;
        }
        if i != len(es6Formatted) {
            fix := strconv.FormatFloat(ieeeF64, 'f', 0, 64)
            if fix[i] >= '5' {
                var adj byte = fix[i - 1] + 1
                es6Formatted = fix[:i - 1] + string(adj) + es6Formatted[i:]
            }
        }
    }
    return sign + es6Formatted, nil
}

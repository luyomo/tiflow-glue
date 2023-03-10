// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"sort"
)

// Range is an interval with a payload.
type Range struct {
	Start   []byte
	End     []byte
	Payload interface{}
}

// Covering represents a non-overlapping, but possibly non-contiguous, set of
// intervals.
type Covering []Range

var _ sort.Interface = Covering{}

func (c Covering) Len() int      { return len(c) }
func (c Covering) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c Covering) Less(i, j int) bool {
	if cmp := bytes.Compare(c[i].Start, c[j].Start); cmp != 0 {
		return cmp < 0
	}
	return bytes.Compare(c[i].End, c[j].End) < 0
}

// OverlapCoveringMerge returns the set of intervals covering every range in the
// input such that no output range crosses an input endpoint. The payloads are
// returned as a `[]interface{}` and in the same order as they are in coverings.
//
// Example:
//
//	covering 1: [1, 2) -> 'a', [3, 4) -> 'b', [6, 7) -> 'c'
//	covering 2: [1, 5) -> 'd'
//	output: [1, 2) -> 'ad', [2, 3) -> `d`, [3, 4) -> 'bd', [4, 5) -> 'd', [6, 7) -> 'c'
//
// The input is mutated (sorted). It is also assumed (and not checked) to be
// valid (e.g. non-overlapping intervals in each covering).
func OverlapCoveringMerge(coverings []Covering) []Range {
	for _, covering := range coverings {
		sort.Sort(covering)
	}
	var ret []Range
	var previousEndKey []byte
	for {
		// Find the start key of the next range. It will either be the end key
		// of the range just added to the output or the minimum start key
		// remaining in the coverings (if there is a gap).
		var startKey []byte
		startKeySet := false
		for _, covering := range coverings {
			if len(covering) == 0 {
				continue
			}
			if !startKeySet || bytes.Compare(covering[0].Start, startKey) < 0 {
				startKey = covering[0].Start
				startKeySet = true
			}
		}
		if !startKeySet {
			break
		}
		if bytes.Compare(startKey, previousEndKey) < 0 {
			startKey = previousEndKey
		}

		// Find the end key of the next range. It's the minimum of all end keys
		// of ranges that intersect the start and all start keys of ranges after
		// the end key of the range just added to the output.
		var endKey []byte
		endKeySet := false
		for _, covering := range coverings {
			if len(covering) == 0 {
				continue
			}

			if bytes.Compare(covering[0].Start, startKey) > 0 {
				if !endKeySet || bytes.Compare(covering[0].Start, endKey) < 0 {
					endKey = covering[0].Start
					endKeySet = true
				}
			}
			if !endKeySet || bytes.Compare(covering[0].End, endKey) < 0 {
				endKey = covering[0].End
				endKeySet = true
			}
		}

		// Collect all payloads of ranges that intersect the start and end keys
		// just selected. Also trim any ranges with an end key <= the one just
		// selected, they will not be output after this.
		var payloads []interface{}
		for i := range coverings {
			// Because of how we chose startKey and endKey, we know that
			// coverings[i][0].End >= endKey and that coverings[i][0].Start is
			// either <= startKey or >= endKey.

			for len(coverings[i]) > 0 {
				if bytes.Compare(coverings[i][0].Start, startKey) > 0 {
					break
				}
				payloads = append(payloads, coverings[i][0].Payload)
				if !bytes.Equal(coverings[i][0].End, endKey) {
					break
				}
				coverings[i] = coverings[i][1:]
			}
		}

		ret = append(ret, Range{
			Start:   startKey,
			End:     endKey,
			Payload: payloads,
		})
		previousEndKey = endKey
	}

	return ret
}

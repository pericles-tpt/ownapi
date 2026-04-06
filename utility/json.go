package utility

import (
	"errors"
	"fmt"
)

type JSONProp struct {
	NestedKey []KeyPart `json:"nested_key"` // e.g. {"name": {"first": "John", "last": "Smith"}} -> ["name", "first"]
	Match     *any      `json:"value"`
	OutputKey *string   `json:"output_key"`
}

type KeyPart struct {
	Key   string `json:"key"`
	Index *uint  `json:"index"`
}

// Match key then add to traverse list
// Match key AND value, extract

func TraverseJSONExtractValues(json any, traverseExtractPlan []JSONProp) (map[string]any, error) {
	var (
		extractedValues    = map[string]any{}
		numExtractedValues = 0
	)
	for i, n := range traverseExtractPlan {
		if len(n.NestedKey) == 0 {
			return extractedValues, fmt.Errorf("invalid node at index %d in `traversePlan`, key must be non-empty", i)
		}
		if n.OutputKey != nil {
			numExtractedValues++
		}
	}
	extractedValues = make(map[string]any, numExtractedValues)

	var (
		extractedIdx int

		jsonTraverseList = []any{json}
		jsonTraverseIdx  int

		lastIsObj bool
		lastIsArr bool

		currObj map[string]any
		currArr []any
	)
	for i := 0; i < len(traverseExtractPlan); i++ {
		var (
			isArr = lastIsArr
			isObj = lastIsObj
		)
		lastIsArr = false
		lastIsObj = false

		pn := traverseExtractPlan[i]

		if jsonTraverseIdx >= len(jsonTraverseList) {
			return extractedValues, errors.New("invalid `traversePlan`, reached the end of the JSON before completing the plan")
		}
		curr := jsonTraverseList[jsonTraverseIdx]

		if !isObj {
			currObj, isObj = curr.(map[string]any)
		}
		if isObj {
			if maybeVal, found := peekVal(currObj, pn.NestedKey); found {
				if pn.Match != nil && maybeVal != *pn.Match {
					jsonTraverseIdx++
					i--
					continue
				}

				if pn.OutputKey != nil {
					extractedValues[*pn.OutputKey] = maybeVal
					extractedIdx++
				} else {
					var (
						isValObj, isValArr bool
					)
					currObj, isValObj = maybeVal.(map[string]any)
					currArr, isValArr = maybeVal.([]any)
					if isValObj || isValArr {
						jsonTraverseList = append(jsonTraverseList, maybeVal)
						jsonTraverseIdx++
					}
					lastIsArr = isValArr
					lastIsObj = isValObj
				}
			} else {
				jsonTraverseIdx++
			}
			continue
		}

		if !isArr {
			currArr, isArr = curr.([]any)
		}
		if isArr {

			// var foundMatch bool
			// for _, elem := range currArr {
			// 	maybeMatch := peekNode(elem, pn)
			// 	if maybeMatch != nil {
			// 		extractedValues[extractedIdx] = elem
			// 		extractedIdx++
			// 		foundMatch = true
			// 		break
			// 	}
			// }

			// if !foundMatch {
			// }
			// TODO: Figure out something better than the line below, look at what I started on above
			jsonTraverseList = append(jsonTraverseList, currArr...)
			jsonTraverseIdx++
			i--
		}
	}

	return extractedValues, nil
}

func peekVal(node any, key []KeyPart) (any, bool) {
	var maybeVal any
	for _, k := range key {
		if k.Index != nil {
			var (
				arr []any
				ok  bool
			)
			if arr, ok = node.([]any); !ok {
				return maybeVal, false
			}

			if *k.Index >= uint(len(arr)) {
				return maybeVal, false
			}

			maybeVal = arr[*k.Index]
			node = arr[*k.Index]
			continue
		}

		var (
			obj map[string]any
			ok  bool
		)
		if len(k.Key) == 0 {
			return maybeVal, false
		}

		if obj, ok = node.(map[string]any); !ok {
			return maybeVal, false
		}

		if maybeVal, ok = obj[k.Key]; !ok {
			return maybeVal, false
		}

		node = maybeVal
	}
	return maybeVal, true
}

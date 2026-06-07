# Functions
Add .go files to this folder to dynamically add functionality to `ownapi`. A `main.go` file has been added but will be ignored for code generation, you can use it to test your functions.

## How it works
1. Create a `.go` file other than `main.go`, this can exist at the root or one level deep in a folder
2. Make sure the package line is: `package main` if at the root OR `package [FOLDER NAME]` if nested
3. Start writing your go code:

   a. Exported functions will be modified by code generation and added to `ownapi`, private functions will be included as-is
      - TIP: You can use private functions to write less limited code, the rules that apply to public functions don't apply to private functions. But, you still can't import libraries not in the whitelist

   b. Comments directly above functions will be included in `ownapi`, all other comments will be omitted

   c. `const` and `var` declarations will be included as-is

   d. PUBLIC FUNCTIONS: Input parameters will be converted to an `[]any` by ownapi, and type assertions will be auto-generated

   e. PUBLIC FUNCTIONS: Return values will be converted to an `([]any, error)`, if your return arguments include one or more error(s), the last error will be return as the 2nd return argument

   f. PUBLIC FUNCTIONS: MUST only use a subset of types (or 1D arrays of them) for the input parameters and return types, these are:

   		"string", "int", "uint", "rune", "byte", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "bool", "time.Time", "error"

   g. Only libraries from the following whitelist can be used:

   		"fmt", "strings", "math", "time", "errors"

5. As long as your code is valid go and follows the above rules, it'll be made available at runtime in `ownapi`
6. You can also modify the code here at runtime and it'll typically be reloaded in < 3s. If the code is invalid you'll see a log in STDOUT and it'll revert to the previous valid plugin
7. You can use the included `main.go` to test your functions, anything in that file will be ignored for code generation. To run `main()` use this command (from root, to run from another path just update the path accordingly):
```
go run ./user_functions/source/main.go
```

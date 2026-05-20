# Functions
Add .go files to this folder to dynamically add functionality to `ownapi`. A `main.go` file has been added but will be ignored, you can use it to test your functions.

## How it works
1. Create a `.go` file other than `main.go`
2. Start writing your go code
    a. Exported functions will be dynamically added to `ownapi`, private functions will be inlined
    b. Comments above functions will be included as documentation in `ownapi`
    c. `const` and `var` declared at the top of the file will be inlined
    d. Input parameters will be converted to an `[]any` by ownapi, and type assertions will be auto-generated
    e. A `propMap map[string]any` will also be included in the parameters
    e. Return values will be stored in the `propMap` as `r0`-`rN`, the only return value in the generated function is `error`
    f. Custom `types` will be ignored, as the only valid types for these functions are go's primitive and standard library types
3. As long as your code is valid go and only uses the supported imports, it'll be imported

## Supported Imports
Not all libraries are supported, any file that imports a package NOT in this list will show a warning and the function won't be added. Supported imports are:
- fmt
- math
- strings
- time

# Static M
Mock services using static files

## File headers
```
path: /json    # the path to match against
method: GET    # the http method to match against (blank/unset = any)
# if the file is a go http/template file or raw text
# when set to true you will have access to named wildcard values from inside the template
template: true 
```

## Wildcards
### Path
- **`*`**: single path segment
- **`*:[name]`**: single path segment with value capture
- **`**`**: greedy path segment (can match one or more path segments)
- **`**:[name]`**: greedy path segment with value capture

## Query String
Query wild cards are named by default, they will use the name of the query param itself
- **`*`**: single value
- **`*:[name]`**: single value with name override
- **`**`**: greedy, this will match one or more values (`?duplicate_key=one&duplicate_key=two` will produce the string `one, two`)
- **`**:[name]`**: greedy with name override

## TODO:
- [ ] validate paths in file headers before setup
- [ ] watch documents directory for changes
- [ ] better logging

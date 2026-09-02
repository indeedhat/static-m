# Static M
Mock services using static files

## Usage
```console
StaticM
Serve static files based on wildcard paths for mocking external servers

Options:
  -port string
    	The port to listen on (default "8080")
  -root string
    	The directory to search for documents within. (default "./documents/")

Flags:
  -watch
    	Watch the directory for file changes.
    	This will parse changes on each request, its wasteful but adequate for a tool like this.
  -v
    	Print verbose output
```

## File headers
```
# Only the path field is required, all other fields are optional

path: /json    # the path to match against
method: GET    # the http method to match against (blank/unset = any)
# if the file is a go http/template file or raw text
# when set to true you will have access to named wildcard values from inside the template
template: true
mime: application/json # the mime type to be returned, if not set this will be auto detected from the file extension
response:
    code: 200 # overwrite the default respnose code
    # map of headers to be returned with the response
    headers:
        x-some-header: Some Value
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
- [x] watch documents directory for changes
- [x] better logging
- [x] auto detect mime types for response header
- [x] allow defining response headers

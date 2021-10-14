# Loader
HTTP Load tester written in Go

# Feature
- Suppress bias of connection destinations by regular name resolution 
- Linear increase to target Request/s

# Usage
```
$ ./loader [Target URL] [Target Request/s] [Seconds to target Request/s]
```
  
## Example
1. Reach 100 Request/s in 3600 seconds.  
```
$ ./loader http://localhost 100 3600
```

# Loader
HTTP Load tester written in Go

# Feature
- Suppress bias of connection destinations by regular name resolution 
- Linear increase to target Request/s

# Usage
```
$ ./loader [Target URL] [Target Request/s] [Time to target Request/s] [Time to exit]
```
  
## Example
1. Reach 100 Request/s in 3600 seconds and wait 3600 seconds to finish.  
```
$ ./loader http://localhost 100 1h 1h
```

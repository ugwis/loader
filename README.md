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

# Traffic Pattern
## linear
![linear](https://user-images.githubusercontent.com/914815/141235700-e44994f9-3e7b-49f7-9d0f-7c97c7c77af7.png)

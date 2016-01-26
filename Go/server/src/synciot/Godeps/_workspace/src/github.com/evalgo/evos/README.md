# evos

## LICENSE
freebsd
## DESCRIPTION
This golang pkg was created to handle some missing functions in the default os golang pkg.
## INSTALL
```sh
	go get evalgo.org/evos
```
## USAGE
First you have to build a rest service and use the evlog pkg
```go
	package {pkgname}
	import(
		"..."
		"evalgo.org/evos"
		"..."
	)

	func EVOsSample(){
	     // ... do something ...
	     exists,err := evos.Exists("/path/to/folder/or/file")
	     if err != nil{
	     	// handle error
	     }
	     if !exists{
	     	// ... do something ...
	     }
	     // ... do something ...
	}
```

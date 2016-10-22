* run `go build air.go`
* Create a script filter node in alfred, as a bash script
* Paste this as the script's content ( to get a list of filterable shows in alfred ):

```
/path_to_air_binary/air path_to_airdates.html list_shows
```
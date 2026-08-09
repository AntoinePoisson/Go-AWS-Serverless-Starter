# The README GIF, recorded by `make demo`. That target clones the repo into a
# throwaway directory first, because the first command below rewrites the
# identity of the checkout it runs in.
#
# The long waits are hidden rather than sped up, there is nothing to watch in
# `go mod tidy` or `npm install`.

Output docs/demo.gif

Require make
Require curl
Require jq
Require docker

Set Shell bash
Set Theme "Catppuccin Mocha"
Set FontSize 16
Set Width 1180
Set Height 620
Set Padding 24
Set TypingSpeed 25ms
Set Framerate 20
# the default is nowhere near enough on a cold cache
Set WaitTimeout 15m

# one command turns the template into a project
Type "./initialize.sh --module github.com/acme/orders-api --service orders-api --reset-releases"
Sleep 700ms
Enter
Sleep 5s

Hide
Wait
Type "clear"
Enter
Show
Sleep 500ms

# runs as a plain HTTP server, no emulator
Type "make run-api &"
Enter
Wait+Screen@5m /listening/
Sleep 2s

Type `curl -s -X POST localhost:8080/items -H 'X-Api-Key: local-dev-key' -d '{"name":"first item","tags":["demo"]}' | jq`
Sleep 400ms
Enter
Sleep 3s

Type "curl -s localhost:8080/items -H 'X-Api-Key: local-dev-key' | jq .items"
Sleep 400ms
Enter
Sleep 4s

# leave nothing running behind the recording
Hide
Type "kill %1"
Enter
Sleep 1s

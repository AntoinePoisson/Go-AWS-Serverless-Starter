# The GIF at the top of the README, recorded by `make demo`.
#
# That target clones the repository into a throwaway directory first, because
# the very first command below rewrites the identity of the checkout it runs in
# - pointed at this one, it would rename the template itself.
#
# The long waits are hidden rather than sped up: `go mod tidy`, `npm install`
# and the DynamoDB container have nothing worth watching. What is left is what
# the README promises - one command to own the template, one to run it, two to
# use it.

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
# What follows installs a Go and a Node toolchain worth of dependencies; on a
# cold cache the default timeout is nowhere near enough.
Set WaitTimeout 15m

# One command turns the template into a project.
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

# It runs as an ordinary HTTP server, no emulator involved.
Type "make run-api &"
Enter
Wait+Screen@5m /listening/
Sleep 2s

# And it answers.
Type `curl -s -X POST localhost:8080/items -H 'X-Api-Key: local-dev-key' -d '{"name":"first item","tags":["demo"]}' | jq`
Sleep 400ms
Enter
Sleep 3s

Type "curl -s localhost:8080/items -H 'X-Api-Key: local-dev-key' | jq .items"
Sleep 400ms
Enter
Sleep 4s

# Nothing is left running behind the recording.
Hide
Type "kill %1"
Enter
Sleep 1s

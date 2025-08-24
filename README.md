# Post office
Basic SMTP server written in Go for [Coding Challenges](https://codingchallenges.fyi/challenges/challenge-smtp/).

## Running

This assumes that you have `go` and `make` installed already.
You can compile the program with:
```
make
```
and run the program with:
```
sudo ./main
```

The `sudo` prefix is needed since the server/client interact with the `25` SMTP port, which requires `root` access.
At this point, the client and the server in the program should begin interacting with one another.

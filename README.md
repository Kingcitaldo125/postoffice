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

## References
https://codingchallenges.fyi/challenges/challenge-smtp/
https://en.wikipedia.org/wiki/Simple_Mail_Transfer_Protocol#SMTP_transport_example

## LICENSE
Post office is licensed under the GNU General Public License 3.0. See the [license file](https://github.com/Kingcitaldo125/postoffice#GPL-3.0-1-ov-file) for more details.

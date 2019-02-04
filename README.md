MarvelAPIDemo with Go
---------------------

Simple Marvel Demo API usage with Go from a console program.

Using the following third party libraries:
- github.com/jroimartin/gocui
- github.com/nfnt/resize
- github.com/qeesung/image2ascii/convert

A basic program.
Make sure to edit `PUBLICKEY` and `PRIVATEKEY` constants in `main.go`file before compile, with your developer's Marvel API Credentials.
Or get it one at https://developer.marvel.com/account

Demo GIF:
![image](https://lh3.googleusercontent.com/ajm6Is6z06Uqns_rzoVsWSyj5jLkMFJr8kFJCi5pKn5B5d4YyZWSAcLsUnQSxeyQyk_3szGdI3ufkQ=w962-h718-no)

Works on Windows and Linux, fails on WSL due to its network limitations.
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
![image](https://lh3.googleusercontent.com/6gOnd_Qs2NozdRibXxsZQuQMQetyxr6T1ZZgW4bhA5vU6tFitbL7eW02HXyA_MfITBBgyxcjnv9ngg=w1366-h728-no)

Works on Windows and Linux, fails on WSL due to its network limitations.
# Installation
Install golang using https://golang.org/doc/install

## Installing Go Dependencies
```
go get
```
## Installing Air
Air allows for hot reload of code
### macOS
```sh
curl -fLo ~/.air \
    https://raw.githubusercontent.com/cosmtrek/air/master/bin/darwin/air
chmod +x ~/.air
```
### Linux
```sh
curl -fLo ~/.air \
    https://raw.githubusercontent.com/cosmtrek/air/master/bin/linux/air
chmod +x ~/.air
```

### Windows
```sh
curl -fLo ~/.air.exe \
    https://raw.githubusercontent.com/cosmtrek/air/master/bin/windows/air.exe
```

# Running the Code
The command air will build the code and run it.
```
air
```
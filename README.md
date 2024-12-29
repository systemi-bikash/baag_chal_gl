# Baag-chal
Ensure you have a proper noice go.mod file .It's the daddyO file for your project to fetch relevant imports 
```
go get github.com/go-gl/gl/v2.1/gl
go get github.com/go-gl/glfw/v3.3/glfw
go get "github.com/golang/freetype"
```
Build & run:
```
go build
go run .
```

## for linux you may require these [ubuntu] 
```
 sudo apt install build-essential pkg-config libgl1-mesa-dev libx11-dev
 sudo apt install libxxf86vm-dev
 sudo apt install libxi-dev
 sudo apt install libxinerama-dev
 sudo apt install libxrandr-dev
 sudo apt install libxcursor-dev
```

make sure to setup 
```
 export PKG_CONFIG_PATH=/path/to/gl.pc:$PKG_CONFIG_PATH
```


Enjoy !!! ( ´∀｀ )

Hard but enjoying www 



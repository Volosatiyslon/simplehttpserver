## Simple http server
### about
simple app to share local files (only files) through http. Auto generated index.html file allows you just start and use. App filter out self file, any directory and hiden files (tested on linux)

 ### usage
 + download a binary from _executable_ folder (or download sources and compile them)
 + put binary to folder, you want to share
 + became shure that to going to run binary under  account, permitted to bind a port
 + server will write to log all accepted network connection

 ### flags
 + use _a_ flag to configure listened address and port (default is 127.0.0.1:8080)\
 ``` simplehttpserver -a 127.0.0.1:8080 ```\
 + use _u_ flag to enable /upload route handler and generate upload form index.html\
 ``` simplehttpserver -u ```\
 by default route /upload is ignored. Uploaded files will be available to download after server restart (will be fixed soon)
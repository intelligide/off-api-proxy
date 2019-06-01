# off-api-proxy
Open Food Facts Api Proxy (for caching, filter &amp; batching)

## Building 

1. Install the prerequisites.
   - Go (We recommend always using the latest stable version)
   - Git

2. Open a terminal.

3. On Unix System:
   ```bash
   # Pick a place for your source.
   $ mkdir -p ~/dev
   $ cd ~/dev
    
   # Grab the code.
   $ git clone https://github.com/intelligide/off-api-proxy.git
    
   # Now we have the source. Time to build!
   $ cd off-api-proxy
    
   # You should be inside ~/dev/off-api-proxy right now.
   $ go run build.go
   ```
   
   On Windows `cmd`:
   ```od
   
   # Pick a place for your source.
   > mkdir %USERPROFILE%\dev
   > cd %USERPROFILE%\dev
   
   # Grab the code.
   > git clone https://github.com/intelligide/off-api-proxy.git
   
   # Now we have the source. Time to build!
   > cd off-api-proxy
   > go run build.go
   ```
   
### Build Options

- `go run build.go install`
  Installs binaries in ./bin (default command, this is what happens 
  when build.go is run without any commands or parameters).
  
- `go run build.go build {target}`
  Builds just the named target, or or all by default, to the `build`
  directory. Use this when cross compiling, with parameters for what
  to cross compile to: `go run build.go -goos linux -goarch 386 build`.
  
- `go run build.go test`
  Runs the tests.
  
- `go run build.go tar`
  Package in a tar.gz dist file in the current directory. Assumes
  a Unixy build.
  
- `go run build.go zip`
  Package in a zip dist file in the current directory. Assumes
  a Windows build.
